package logdb

import (
	"root/core/db"
	"root/core/log"
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"sync"
)

type RMBModel struct {
	ID          uint32 `gorm:"column:log_ID; primary_key; auto_increment"`                                 //序号
	AccountID   uint32 `gorm:"column:log_AccountID; primary_key; TYPE:int unsigned; not null"`             //玩家AccountID
	ChangeValue int64  `gorm:"column:log_ChangeValue; not null; default 0"`                                //改变值
	Value       int64  `gorm:"column:log_Value; not null; default 0"`                                      //改变后剩余值
	Index       uint8  `gorm:"column:log_Index; not null; default 0"`                                      //下注区域
	Operate     uint8  `gorm:"column:log_Operate; not null; default 0"`                                    //改变原因
	Time        string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"` //改变时间
	RoomID      uint32 `gorm:"column:log_RoomID; not null; default 0"`                                     //房间ID
	GameType    uint8  `gorm:"column:log_GameType; not null; default 0"`
	logPortion  int
}

func (self *RMBModel) Reset()                  { self = &RMBModel{} }
func (self *RMBModel) String() string          { return "" }
func (self *RMBModel) ProtoMessage()           {}
func (self *RMBModel) Portion(i int) *RMBModel { self.logPortion = i; return self }

//自定义表名
func (self *RMBModel) TableName() string {
	//strSuffix := time.Now().Format("2006-01-02")
	//return "log_rmb_" + strSuffix
	return "log_rmb_" + strconv.Itoa(int(self.logPortion))
}

func ShiftRMB() string {
	all := []*RMBModel{}
	logcon := db.GetLog()
	if logcon == nil {
		return "error"
	}

	logcon.Table("log_rmb").Find(&all)

	logc := [10]*gorm.DB{}

	wg := &sync.WaitGroup{}
	channels := [10]chan interface{}{}
	channelse := [10]chan int{}

	for i := 0; i < 10; i++ {
		logc[i] = db.GetLog2()
		channels[i] = make(chan interface{}, 50000)
		channelse[i] = make(chan int)
		exit := false
		exec := func(c chan interface{}, e chan int, index int) {
			wg.Add(1)
			defer func() {
				wg.Done()
			}()

			str_pre := fmt.Sprintf("INSERT INTO log_rmb_%v(log_AccountID, log_ChangeValue, log_Value,log_Index,log_Operate,log_Time,log_RoomID,log_GameType, log_ClubID) VALUES", index)
			sql := str_pre
			count := 0
			for {
				select {
				case new := <-c:
					f := new.(func(string) string)
					sql = f(sql)
					count++
					if count == 1000 {
						count = 0
						sql = strings.TrimRight(sql, ",")
						e := logc[index].Exec(sql)

						if e.Error != nil {
							log.Warnf("数据插入错误:%v error:%v", sql, e.Error.Error())
							return
						}
						//log.Infof("数据转移成功 table:log_rmb_%v sql:%v", index, sql)
						sql = str_pre
					}

					l := len(c)
					if l == 0 && exit {
						if count != 0 {
							sql = strings.TrimRight(sql, ",")
							e := logc[index].Exec(sql)
							if e.Error != nil {
								log.Warnf("数据插入错误:%v error:%v", sql, e.Error.Error())
								return
							}
							//log.Infof("数据转移成功 尾巴 table:log_rmb_%v sql:%v", index, sql)
						}

						log.Infof("%v号表完成 count:%v ", index, count)
						return
					}
				case <-e:
					exit = true
					if len(c) == 0 {
						log.Infof("%v号表完成", index)
						return
					}
					log.Infof("%v号表收到退出信号", index)
				}
			}
		}
		go exec(channels[i], channelse[i], i)
	}

	log.Infof("数据分析...")
	var staticNum [10]int
	total := len(all)
	for _, data := range all {
		splitNum := data.AccountID % 10
		tempd := data
		tempd.ID = 0
		channels[int(splitNum)] <- func(pre string) string {
			tempd.Portion(int(splitNum))
			data := fmt.Sprintf("(%v,%v,%v,%v,%v,'%v',%v,%v),", tempd.AccountID, tempd.ChangeValue, tempd.Value, tempd.Index, tempd.Operate, tempd.Time, tempd.RoomID, tempd.GameType)
			sql := pre + data
			staticNum[splitNum]++
			return sql
		}
	}

	log.Infof("数据分析完成")
	for i := 0; i < 10; i++ {
		channelse[i] <- 1
	}

	wg.Wait()
	for i := 0; i < 10; i++ {
		logc[i].Close()
	}
	return fmt.Sprintf("log_rmb 转移完成 total:%v条数据 static:%v", total, staticNum)
}
