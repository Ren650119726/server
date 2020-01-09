package config

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

type public_type map[string]string

//val 只有两种，int64 string
var global_public_config public_type

// 创建房间参数列表 key:游戏类型 value:参数字符串切片
var global_create_room_param map[uint8][]string

// 启服自动创建房间参数列表 key:游戏类型 value:参数字符串切片
var global_auto_create_room_param map[uint8][]string

var lock sync.Mutex

func init() {
	LoadPublic_Conf()
}

func LoadPublic_Conf() {
	lock.Lock()
	defer lock.Unlock()

	global_public_config = public_type{}
	data, e := ioutil.ReadFile(core.ConfigDir + "public.json")
	if e != nil {
		log.Errorf("public 错误:%v", e.Error())
		return
	}
	err := json.Unmarshal(data, &global_public_config)
	if err != nil {
		log.Errorf(" error %v", err.Error())
		return
	}

	global_auto_create_room_param = make(map[uint8][]string)
	global_auto_create_room_param[common.EGameTypeNIU_NIU.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_5"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_6"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_7"],
	}
	global_auto_create_room_param[common.EGameTypeWUHUA_NIUNIU.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_5"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_6"],
		global_public_config["NN_AUTO_CREATE_ROOM_PARAM_7"],
	}
	global_auto_create_room_param[common.EGameTypeTEN_NIU_NIU.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_5"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_6"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_7"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_8"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_9"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_10"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_11"],
		global_public_config["TNN_AUTO_CREATE_ROOM_PARAM_12"],
	}
	global_auto_create_room_param[common.EGameTypeSAN_GONG.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_5"],
		global_public_config["SG_AUTO_CREATE_ROOM_PARAM_6"],
	}
	global_auto_create_room_param[common.EGameTypeSHI_SAN_SHUI.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["SSS_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["SSS_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["SSS_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["SSS_AUTO_CREATE_ROOM_PARAM_4"],
	}
	global_auto_create_room_param[common.EGameTypeCHE_XUAN.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["CX_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["CX_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["CX_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["CX_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["CX_AUTO_CREATE_ROOM_PARAM_5"],
	}
	global_auto_create_room_param[common.EGameTypeDING_ER_HONG.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["DEH_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["DEH_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["DEH_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["DEH_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["DEH_AUTO_CREATE_ROOM_PARAM_5"],
	}
	global_auto_create_room_param[common.EGameTypeDGK.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_21"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_22"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_23"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_24"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_25"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_26"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_31"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_32"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_33"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_34"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_35"],
		global_public_config["DGK_AUTO_CREATE_ROOM_PARAM_36"],
	}
	global_auto_create_room_param[common.EGameTypeXMMJ.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_21"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_22"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_23"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_24"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_25"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_31"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_32"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_33"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_34"],
		global_public_config["PANDA_AUTO_CREATE_ROOM_PARAM_35"],
	}
	global_auto_create_room_param[common.EGameTypePAO_DE_KUAI.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["PDK_AUTO_CREATE_ROOM_PARAM_31"],
		global_public_config["PDK_AUTO_CREATE_ROOM_PARAM_32"],
		global_public_config["PDK_AUTO_CREATE_ROOM_PARAM_33"],
	}
	global_auto_create_room_param[common.EGameTypePDK_HN.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_21"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_22"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_23"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_24"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_25"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_26"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_27"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_28"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_29"],
		global_public_config["PDK_HN_AUTO_CREATE_ROOM_PARAM_30"],
	}
	global_auto_create_room_param[common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["SSZZ_AUTO_CREATE_ROOM_PARAM_1"],
	}
	global_auto_create_room_param[common.EGameTypeFQZS.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["SSZZ_AUTO_CREATE_ROOM_PARAM_1"],
	}
	global_auto_create_room_param[common.EGameTypeTUI_TONG_ZI.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["TTZ_AUTO_CREATE_ROOM_PARAM_1"],
	}
	global_auto_create_room_param[common.EGameTypeLONG_HU_DOU.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["LHD_AUTO_CREATE_ROOM_PARAM_1"],
	}
	global_auto_create_room_param[common.EGameTypeHONG_HEI_DA_ZHAN.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["R2B_AUTO_CREATE_ROOM_PARAM_1"],
	}
	global_auto_create_room_param[common.EGameTypeHONG_BAO.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["HB_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["HB_AUTO_CREATE_ROOM_PARAM_2"],
	}

	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_6")
	delete(global_public_config, "NN_AUTO_CREATE_ROOM_PARAM_7")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "SG_AUTO_CREATE_ROOM_PARAM_6")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_6")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_7")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_8")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_9")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_10")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_11")
	delete(global_public_config, "TNN_AUTO_CREATE_ROOM_PARAM_12")
	delete(global_public_config, "SSS_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "SSS_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "SSS_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "SSS_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "CX_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "CX_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "CX_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "CX_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "CX_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "DEH_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "DEH_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "DEH_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "DEH_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "DEH_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_21")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_22")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_23")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_24")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_25")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_26")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_31")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_32")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_33")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_34")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_35")
	delete(global_public_config, "DGK_AUTO_CREATE_ROOM_PARAM_36")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_21")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_22")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_23")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_24")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_25")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_31")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_32")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_33")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_34")
	delete(global_public_config, "PANDA_AUTO_CREATE_ROOM_PARAM_35")
	delete(global_public_config, "PDK_AUTO_CREATE_ROOM_PARAM_31")
	delete(global_public_config, "PDK_AUTO_CREATE_ROOM_PARAM_32")
	delete(global_public_config, "PDK_AUTO_CREATE_ROOM_PARAM_33")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_21")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_22")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_23")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_24")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_25")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_26")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_27")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_28")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_29")
	delete(global_public_config, "PDK_HN_AUTO_CREATE_ROOM_PARAM_30")
	delete(global_public_config, "SSZZ_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "TTZ_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "LHD_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "R2B_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "HB_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "HB_AUTO_CREATE_ROOM_PARAM_2")

	// 以下是机器人创建房间参数
	// 手动组装创建房间参数
	global_create_room_param = make(map[uint8][]string)
	global_create_room_param[common.EGameTypeJIN_HUA.Value()] = []string{
		"ERROR_PARAM", // 为了和客户端传来的下标参数匹配, 不用-1;
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_1"],
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_2"],
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_3"],
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_4"],
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_5"],
		global_public_config["JH_AUTO_CREATE_ROOM_PARAM_6"],
	}
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_1")
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_2")
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_3")
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_4")
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_5")
	delete(global_public_config, "JH_AUTO_CREATE_ROOM_PARAM_6")

	log.Info("加载完成public.json")
}

// 获取创建房间参数长度
func GetCreateRoomParamLen(nGameType uint8) uint8 {
	lock.Lock()
	defer lock.Unlock()

	tNode := global_create_room_param[nGameType]
	if tNode != nil {
		return uint8(len(tNode))
	}
	return 0
}

// 获取创建房间参数字符串
func GetCreateRoomParam(nGameType uint8, nMatchType uint8) string {
	lock.Lock()
	defer lock.Unlock()

	tNode := global_create_room_param[nGameType]
	if tNode != nil && nMatchType < uint8(len(tNode)) {
		return tNode[nMatchType]
	}
	return "ERROR_PARAM"
}

// 获取创建房间参数长度
func GetAutoCreateRoomParamLen(nGameType uint8) uint8 {
	lock.Lock()
	defer lock.Unlock()

	tNode := global_auto_create_room_param[nGameType]
	if tNode != nil {
		return uint8(len(tNode))
	}
	return 0
}

// 获取创建房间参数字符串
func GetAutoCreateRoomParam(nGameType uint8, nMatchType uint8) string {
	lock.Lock()
	defer lock.Unlock()

	tNode := global_auto_create_room_param[nGameType]
	if tNode != nil && nMatchType < uint8(len(tNode)) {
		return tNode[nMatchType]
	}
	return "ERROR_PARAM"
}

func GetPublicConfig_Int64(key string) int64 {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	if p == "" {
		log.Errorf("全局配置找不到:[%v]", key)
		return 0
	}

	val, err := strconv.Atoi(p)
	if err != nil {
		log.Errorf("字符串不是整型:%v", p)
		return 0
	}

	return int64(val)
}

func GetPublicConfig_Float32(key string) float32 {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	if p == "" {
		log.Errorf("全局配置找不到:[%v]", key)
		return 0
	}

	val, err := strconv.ParseFloat(p, 32)
	if err != nil {
		log.Errorf("字符串不是浮点型:%v", p)
		return 0
	}

	return float32(val)
}

func GetPublicConfig_String(key string) string {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	return p
}

func GetPublicConfig_Slice(key string) []int {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	valString := p
	ret := make([]int, 0)
	s1 := strings.Split(valString, "*")
	for _, v := range s1 {
		value, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", valString, v)
			return nil
		}
		ret = append(ret, value)
	}
	return ret
}

func GetPublicConfig_Mapi(key string) map[int]int {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	if p == "" {
		ret := make(map[int]int)
		return ret
	}

	ret := utils.SplitConf2Mapii(p)
	return ret
}

// 解析"1#100,2#200,3#300"为[[1,100],[2,200],[3,300]]
func GetPublicConfig_ArrInt64(key string) [][]int64 {
	lock.Lock()
	defer lock.Unlock()
	p, _ := global_public_config[key]
	ret := utils.SplitConf2Arr_ArrInt64(p)
	return ret
}

func GetPublicConfig_MapStrInt(key string) map[string]int {
	lock.Lock()
	defer lock.Unlock()
	p, _ := global_public_config[key]
	ret := utils.SplitConf2MapStrInt(p)
	return ret
}

func GetPublicConfig_MapStrStr(key string) map[string]string {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	if p == "" {
		ret := make(map[string]string)
		return ret
	}

	ret := utils.SplitConf2MapStrStr(p, "|", "*")
	return ret
}

// 配置格式: 0*30|1*40|2*30; 根据概率返回0,1,2之一
// 失败返回: 第二参数默认值
// 随到0的概率30%  随到1概率40  随到2概率30
// 概率总和可以是100% 也可以超过100%
func GetPublicConfig_CalcRandomReturnKey(key string, default_ret int) int {
	lock.Lock()
	defer lock.Unlock()

	p, _ := global_public_config[key]
	if p == "" {
		return default_ret
	}

	val := utils.CalcRandomReturnKey(p, default_ret)
	return val
}

func IsHaveBannedWords(text string) bool {
	lock.Lock()
	defer lock.Unlock()
	p, _ := global_public_config["BANNED_WORDS"]
	ret := strings.Split(p, "|")

	check := strings.ToLower(text)
	for _, words := range ret {
		if strings.Contains(check, words) == true {
			return true
		}
	}
	return false
}

// 第一返回: 是否是测试服务器; true表示是
// 第二返回: 本地IP
// 第三返回: 外网IP
func IsTestServer() (bool, string, string) {
	lock.Lock()
	defer lock.Unlock()

	strLocalIP := utils.GetLocalIP()
	var mTestServerIP map[string]string
	if strConf, isExistList := global_public_config["TEST_SERVER_LIST"]; isExistList == true {
		mTestServerIP = utils.SplitConf2MapStrStr(strConf, "|", "*")
	}
	if mTestServerIP == nil {
		return false, strLocalIP, ""
	}

	if strRealIP, isExist := mTestServerIP[strLocalIP]; isExist == true {
		return true, strLocalIP, strRealIP
	} else {
		return false, strLocalIP, ""
	}
}
