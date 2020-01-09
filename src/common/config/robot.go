package config

import (
	"root/core"
	"root/core/log"
	"root/core/utils"
	"encoding/json"
	"io/ioutil"
	"time"
)

type (
	robotName struct {
		RobotID int    `json:"RobotID"`
		Name    string `json:"Name"`
		HeadURL string `json:"HeadURL"`
		Enable  int    `json:"Enable"`
		Grade   int    `json:"Grade"`
	}

	RobotTime struct {
		StartTime string `json:"StartTime"`
		EndTime   string `json:"EndTime"`
		JH        int    `json:"JH"`
		NN        string `json:"NN"`
		WHNN      string `json:"WHNN"`
		SSS       string `json:"SSS"`
		SSZZ      string `json:"SSZZ"`
		TTZ       string `json:"TTZ"`
		LHD       string `json:"LHD"`
		R2D       string `json:"R2D"`
		HB        string `json:"HB"`
		DGK       string `json:"DGK"`
		XMMJ      string `json:"XMMJ"`
		TNN       string `json:"TNN"`
		FQZS      string `json:"FQZS"`
		PDK_HN    string `json:"PDK_HN"`
		SG        string `json:"SG"`
		SG_WATCH  int    `json:"SG_WATCH"`
	}

	robot_name_type map[int]*robotName
	robot_time_type map[int]*RobotTime
)

var global_robot_name_config robot_name_type
var global_robot_time_config []*RobotTime

func init() {
	LoadRobot_Conf()
}

func GetRobotNameConfig() robot_name_type {
	lock.Lock()
	defer lock.Unlock()

	return global_robot_name_config
}

func GetNowRobotConfig() (*RobotTime, int64) {

	lock.Lock()
	defer lock.Unlock()

	nNowTime := utils.MilliSecondTimeSince1970()
	strPrefix := time.Now().Format("2006-01-02")
	var tConfig *RobotTime
	for _, tNode := range global_robot_time_config {
		nStart := utils.String2UnixStamp(strPrefix + " " + tNode.StartTime)
		nEnd := utils.String2UnixStamp(strPrefix + " " + tNode.EndTime)
		if nStart > nEnd {
			nEnd += 86400000
		}
		if nNowTime >= nStart && nNowTime < nEnd {
			tConfig = tNode
			break
		}
	}
	if tConfig == nil {
		tConfig = global_robot_time_config[0]
	}
	return tConfig, nNowTime
}

func LoadRobot_Conf() {
	lock.Lock()
	defer lock.Unlock()

	global_robot_name_config = robot_name_type{}
	dataName, eName := ioutil.ReadFile(core.ConfigDir + "robotName.json")
	if eName != nil {
		log.Errorf("robot name 错误1:%v", eName.Error())
		return
	}
	errorName := json.Unmarshal(dataName, &global_robot_name_config)
	if errorName != nil {
		log.Errorf("robot name 错误2 %v", errorName.Error())
		return
	}

	// 检测机器人配置中是否有重复配置
	mCheckID := make(map[int]bool)
	//mCheckName := make(map[string]bool)
	//mCheckHead := make(map[string]bool)
	for nID, tNode := range global_robot_name_config {
		if _, isExist := mCheckID[tNode.RobotID]; isExist == true {
			log.Warnf("====================> Robot Config Error, ID:%v  Error RobotID:%v", nID, tNode.RobotID)
		} else {
			mCheckID[tNode.RobotID] = true
		}
		//if _, isExist := mCheckName[tNode.Name]; isExist == true {
		//	log.Warnf("====================> Robot Config Error, ID:%v  Error RobotName:%v", nID, tNode.Name)
		//} else {
		//	mCheckName[tNode.Name] = true
		//}
		//if _, isExist := mCheckHead[tNode.HeadURL]; isExist == true {
		//	log.Warnf("====================> Robot Config Error, ID:%v  Error RobotHeadURL:%v", nID, tNode.HeadURL)
		//} else {
		//	mCheckHead[tNode.HeadURL] = true
		//}
	}

	//////////////////////////////////////////////////////////////////////////////////////////
	temp_robot_time_config := make(map[int]*RobotTime)
	dataTime, eTime := ioutil.ReadFile(core.ConfigDir + "robotTime.json")
	if eTime != nil {
		log.Errorf("robot time 错误1:%v", eTime.Error())
		return
	}
	errorTime := json.Unmarshal(dataTime, &temp_robot_time_config)
	if errorTime != nil {
		log.Errorf("robot time 错误2 %v", errorTime.Error())
		return
	}
	nRobotTimeLen := len(temp_robot_time_config)
	global_robot_time_config = make([]*RobotTime, nRobotTimeLen)
	for nID, tNode := range temp_robot_time_config {
		global_robot_time_config[nID-1] = tNode
	}
	log.Info("加载完成robotTime.json和robotName.json")

}
