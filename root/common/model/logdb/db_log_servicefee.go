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

type ServiceModel struct {
	ID         uint32 `gorm:"column:log_ID; primary_key; auto_increment"`                                 //序号
	AccountID  uint32 `gorm:"column:log_AccountID; primary_key; TYPE:int unsigned; not null"`             //玩家AccountID
	ServiceFee uint32 `gorm:"column:log_ServiceFee; not null"`                                            //改变值
	GameType   uint8  `gorm:"column:log_GameType; not null"`                                              //改变原因
	Time       string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"` //改变时间
	RoomID     uint32 `gorm:"column:log_RoomID; not null"`                                                //房间ID
	logPortion int
}

func (self *ServiceModel) Reset()                      { self = &ServiceModel{} }
func (self *ServiceModel) String() string              { return "" }
func (self *ServiceModel) ProtoMessage()               {}
func (self *ServiceModel) Portion(i int) *ServiceModel { self.logPortion = i; return self }

//自定义表名
func (self *ServiceModel) TableName() string {
	//strSuffix := time.Now().Format("2006-01-02")
	//return "log_servicefee_" + strSuffix
	return "log_servicefee_" + strconv.Itoa(int(self.logPortion))
}

func ShiftServiceFee() string {
	all := []*ServiceModel{}
	logcon := db.GetLog()
	if logcon == nil {
		return "error"
	}

	logcon.Table("log_servicefee").Find(&all)

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

			str_pre := fmt.Sprintf("INSERT INTO log_servicefee_%v(log_AccountID, log_ServiceFee, log_GameType,log_Time,log_RoomID, log_ClubID) VALUES", index)
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
						//log.Infof("数据转移成功 table:log_servicefee_%v sql:%v", index, sql)
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
			data := fmt.Sprintf("(%v,%v,%v,'%v',%v),", tempd.AccountID, tempd.ServiceFee, tempd.GameType, tempd.Time, tempd.RoomID)
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
	return fmt.Sprintf("log_servicefee 转移完成 total:%v条数据 static:%v", total, staticNum)
}
