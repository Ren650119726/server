package logic

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"encoding/json"
	"fmt"
	"math"
	"root/protomsg"
	"regexp"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"time"
)

// 登记兑换支付信息
func (self *Hall) Old_MSGID_REGISTER_EXCHANGE_INFO(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	channel := pack.ReadUInt8()
	exchangeId := pack.ReadString()

	log.Debugf(" 登记兑换信息 acc:%v channel:%v exchange:%v", accountId, channel, exchangeId)
	acc := account.CheckSession(accountId, session)
	if acc == nil {
		log.Errorf("登记兑换信息 找不到acc session :%v", session)
		return
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_REGISTER_EXCHANGE_INFO.UInt16())
	if channel == 0 || channel != 1 && channel != 2 && channel != 3 {
		log.Errorf("登记兑换信息 channel error :%v", channel)
		send.WriteUInt8(1)
		send.WriteUInt8(channel)
		send.WriteString(exchangeId)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 支付账户信息是""
	if exchangeId == "" {
		log.Errorf("登记兑换信息 exchangeId error :%v", exchangeId)
		send.WriteUInt8(2)
		send.WriteUInt8(channel)
		send.WriteString(exchangeId)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if channel == 2 {
		s := struct {
			Name string
			Id   string
		}{}
		json.Unmarshal([]byte(exchangeId), &s)
		if result, _ := regexp.MatchString(`^\d{11}$`, s.Id); !result {
			send.WriteUInt8(3)
			send.WriteUInt8(channel)
			send.WriteString(exchangeId)
			send_tools.Send2Account(send.GetData(), session)
			return
		}

		acc.ZhifuBaoExchangeID = exchangeId // 客户端申请填写支付包信息
	} else if channel == 3 {
		acc.BackExchangeID = exchangeId
	}

	send.WriteUInt8(0)
	send.WriteUInt8(channel)
	send.WriteString(exchangeId)
	send_tools.Send2Account(send.GetData(), session)
}

// 兑现请求
func (self *Hall) Old_MSGID_EXCHANGE_RMB(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	amount := pack.ReadUInt32()
	exchangeChannel := pack.ReadUInt8()

	acc := account.CheckSession(nAccountID, session)
	if acc == nil {
		log.Errorf("Can't find Account , session:%v", session)
		return
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_EXCHANGE_RMB.UInt16())

	curRMB := acc.GetMoney()
	if curRMB < uint64(amount) {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	EXCHANGE_LIMIT := config.GetPublicConfig_Slice("EXCHANGE_LIMIT")
	if int64(amount) < int64(EXCHANGE_LIMIT[0]) || int64(amount) > int64(EXCHANGE_LIMIT[1]) {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	userExchangeId := ""
	if exchangeChannel == 2 {
		if acc.ZhifuBaoExchangeID == "" {
			send.WriteUInt8(2)
			send_tools.Send2Account(send.GetData(), session)
			return
		}
		EXCHANGE_OPEN_ZFB := config.GetPublicConfig_Int64("EXCHANGE_OPEN_ZFB")
		if EXCHANGE_OPEN_ZFB == 0 {
			send.WriteUInt8(4)
			send_tools.Send2Account(send.GetData(), session)
			return
		}
		userExchangeId = acc.ZhifuBaoExchangeID
	} else if exchangeChannel == 3 {
		if acc.BackExchangeID == "" {
			send.WriteUInt8(2)
			send_tools.Send2Account(send.GetData(), session)
			return
		}
		userExchangeId = acc.BackExchangeID
	}

	// 计算到帐金额; 星期六免手续费
	EXCHANGE_SCALE := config.GetPublicConfig_Mapi("EXCHANGE_SCALE")
	t := time.Now()
	nWeekDay := int(t.Weekday())
	nExchangeScale, isExist := EXCHANGE_SCALE[nWeekDay]
	if isExist == false {
		send.WriteUInt8(5)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	orderId := fmt.Sprintf("%v_%v", acc.AccountId, utils.MilliSecondTimeSince1970())
	arrival := amount * uint32(nExchangeScale) / 100
	buildTime := utils.DateString()
	order := &protomsg.ExchangeOrder{
		Order:          orderId,
		AccountId:      acc.AccountId,
		Amount:         amount,
		Arrival:        arrival,
		CreationTime:   buildTime,
		ExchangeCannel: uint32(exchangeChannel),
		ExchangeID:     userExchangeId,
	}

	log.Infof("兑现申请:%v 申请金额:%v 到帐金额:%v 兑换信息:%v", acc.AccountId, amount, arrival, exchangeChannel)

	if exchangeChannel == 2 { // 支付宝方式类型2
		// 目前支付宝是代付方式; 默认自动审核
		order.ExchangeStatus = common.EXCHANGE_STATE_AUTO_REVIEW.Value()
		sJson, jerr := json.Marshal(order)
		if jerr != nil {
			send.WriteUInt8(4)
			send_tools.Send2Account(send.GetData(), session)
			return
		}
		strJson := string(sJson)
		isTestServer, _, _ := config.IsTestServer()
		if isTestServer == false {
			HallMgr.sExchangeChan <- strJson
		} else {
			// 测试服务器下支付宝通道默认人工审核
			order.ExchangeStatus = common.EXCHANGE_STATE_MANUAL_REVIEW.Value()
		}
	} else if exchangeChannel == 3 { // 银行卡方式类型3
		// 银行卡是默认人工审核
		order.ExchangeStatus = common.EXCHANGE_STATE_MANUAL_REVIEW.Value()
	}

	if _, isExist := account.AccountMgr.ExchangeOrder[orderId]; isExist == false {
		account.AccountMgr.ExchangeOrder[orderId] = order
	}

	send_tools.Send2DB(account.AccountMgr.StampNum(), protomsg.MSGID_HG_INSERT_EXCHANGE_ORDER.UInt16(), &protomsg.INSERT_EXCHANGE_ORDER{Order: order}, true)
	log.Infof("SQL_INST: INSERT INTO gd_exchangeorder (gd_Order, gd_AccountID, gd_Amount, gd_Arrival, gd_CreationTime, gd_ExchangeCannel, gd_ExchangeID, gd_ExchangeStatus, gd_ExchangeSuccessTime, gd_HandlerInfo) VALUES ('%v', %v, %v, %v, '%v', %v, '%v', %v, '%v', '%v');",
		order.Order, order.AccountId, order.Amount, order.Arrival, order.CreationTime, order.ExchangeCannel, order.ExchangeID, order.ExchangeStatus, order.ExchangeSuccessTime, order.HandlerInfo)

	// 后扣钱, 避免json转码失败
	acc.AddMoney(-int64(amount), common.EOperateType_EXCHANGE)

	send.WriteUInt8(0)
	send.WriteString(orderId)
	send.WriteUInt32(amount)
	send.WriteUInt32(arrival)
	send.WriteString(buildTime)
	send.WriteUInt8(exchangeChannel)
	send.WriteString(userExchangeId)
	send.WriteUInt8(0)
	send.WriteString("")
	send_tools.Send2Account(send.GetData(), session)

	// 添加到历史订单列表中, 只存9条数据
	EXCHANGE_ORDER_MAX_COUNT := config.GetPublicConfig_Int64("EXCHANGE_ORDER_MAX_COUNT")
	acc.ExchangeOrderList = append(acc.ExchangeOrderList, order)
	if len(acc.ExchangeOrderList) > int(EXCHANGE_ORDER_MAX_COUNT) {
		acc.ExchangeOrderList = acc.ExchangeOrderList[1:]
	}

	// 广播有玩家提现小喇叭
	accString := fmt.Sprintf("%v**", uint64(math.Floor(float64(acc.AccountId/100))))
	combinationOfString := fmt.Sprintf(config.GetPublicConfig_String("EXCHANGE_NOTIFY"), accString, amount/100)
	broadcast := packet.NewPacket(nil)
	broadcast.SetMsgID(protomsg.Old_MSGID_SEND_SPEAKERS.UInt16())
	broadcast.WriteUInt8(0)
	broadcast.WriteString(combinationOfString)
	account.AccountMgr.SendBroadcast(broadcast, 2)

	//// 更新提现通知时间
	//sxy := utils.SplitConf2ArrInt32(config.GetPublicConfig_String("EXCHANGE_RANDOM_RANGE"), "*")
	//nextLieTime := utils.Randx_y(int(sxy[0]), int(sxy[1]))
	//HallMgr.nLieExchangeNotify = utils.SecondTimeSince1970() + int64(nextLieTime)
}

// 兑现订单列表
func (self *Hall) Old_MSGID_EXCHANGE_ORDER_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	//log.Debugf("请求兑换订单列表 accountId:%v", accountId)
	acc := account.CheckSession(accountId, session)
	if acc == nil {
		log.Errorf("can find acc :%v", accountId)
		return
	}

	count := len(acc.ExchangeOrderList)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_EXCHANGE_ORDER_LIST.UInt16())
	tSend.WriteUInt16(uint16(count))
	for _, order := range acc.ExchangeOrderList {
		tSend.WriteString(order.Order)
		tSend.WriteUInt32(order.Amount)
		tSend.WriteUInt32(order.Arrival)
		tSend.WriteString(order.CreationTime)
		tSend.WriteUInt8(uint8(order.ExchangeCannel))
		tSend.WriteString(order.ExchangeID)
		tSend.WriteUInt8(uint8(order.ExchangeStatus))
		tSend.WriteString(order.ExchangeSuccessTime)
	}
	send_tools.Send2Account(tSend.GetData(), session)
}
