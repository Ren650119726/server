package db

import (
	"database/sql"
	"fmt"
	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"time"
)

var (
	game_inst *gorm.DB
	game_log  *gorm.DB
	game_web  *gorm.DB

	libs []string
	addr string

	inst_time int64
	log_time  int64
)

const timeout_ses = 60000

// 初始化
func init() {
	libs = beego.AppConfig.Strings(core.Appname + "::mysqllibs")
	addr = beego.AppConfig.String(core.Appname + "::mysql")
	if addr == ""{
		return
	}
	mysql, err := sql.Open("mysql", fmt.Sprintf(addr, ""))
	defer mysql.Close()
	if err != nil {
		log.Errorf("mysql 连接失败:%v", err.Error())
		return
	}
	log.Infof(colorized.Gray("%v mysql 连接成功"), addr)

	buildDataBase_command := "CREATE DATABASE `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
	for _, dbName := range libs {
		_, e := mysql.Exec(fmt.Sprintf(buildDataBase_command, dbName))
		if e == nil {
			log.Infof(colorized.Gray("%v 库创建成功"), dbName)
		} else {
			log.Warnf(":%v", e)
		}
	}
}

func connectInst() *gorm.DB {
	if len(libs) < 2 {
		log.Errorf("数据库配置错误")
		return nil
	}

	libs = beego.AppConfig.Strings(core.Appname + "::mysqllibs")
	addr = beego.AppConfig.String(core.Appname + "::mysql")
	// 获取配置
	maxidleconns := beego.AppConfig.DefaultInt(core.Appname+"::maxidleconns", 0)
	maxopenconns := beego.AppConfig.DefaultInt(core.Appname+"::maxopenconns", 0)

	lib_inst := fmt.Sprintf(addr, libs[0])
	game_inst = api_open(lib_inst, maxopenconns, maxidleconns)
	if game_inst == nil {
		log.Fatalf("无法连接实例数据库: %v", lib_inst)
		return nil
	}

	return game_inst
}
func connectLog() *gorm.DB {
	if len(libs) < 2 {
		log.Errorf("数据库配置错误")
		return nil
	}

	libs = beego.AppConfig.Strings(core.Appname + "::mysqllibs")
	addr = beego.AppConfig.String(core.Appname + "::mysql")
	// 获取配置
	maxidleconns := beego.AppConfig.DefaultInt(core.Appname+"::maxidleconns", 0)
	maxopenconns := beego.AppConfig.DefaultInt(core.Appname+"::maxopenconns", 0)

	lib_log := fmt.Sprintf(addr, libs[1])
	glog := api_open(lib_log, maxopenconns, maxidleconns)
	if glog == nil {
		log.Fatalf("无法连接日志数据库: %v", lib_log)
		return nil
	}
	return glog
}
func connectWeb() *gorm.DB {
	if len(libs) < 2 {
		log.Errorf("数据库配置错误")
		return nil
	}

	libs = beego.AppConfig.Strings(core.Appname + "::mysqllibs")
	addr = beego.AppConfig.String(core.Appname + "::mysql")
	// 获取配置
	maxidleconns := beego.AppConfig.DefaultInt(core.Appname+"::maxidleconns", 0)
	maxopenconns := beego.AppConfig.DefaultInt(core.Appname+"::maxopenconns", 0)

	if len(libs) != 3 {
		log.Error("app.config mysqllibs配置错误缺少game_web")
	}
	lib_web := fmt.Sprintf(addr, libs[2])
	glog := api_open(lib_web, maxopenconns, maxidleconns)
	if glog == nil {
		log.Fatalf("无法连接日志数据库: %v", lib_web)
		return nil
	}
	return glog
}

// 打开一个数据库
func api_open(addr string, maxopenconns, maxidleconns int) *gorm.DB {
	if addr == "" {
		log.Error("add is nil")
		return nil
	}

	db, err := gorm.Open("mysql", addr)
	if err != nil {
		log.Error("open mysql error", err.Error(), addr)
		return nil
	}

	db.DB().SetConnMaxLifetime(time.Second * 5)
	db.DB().SetMaxIdleConns(maxidleconns)
	db.DB().SetMaxOpenConns(maxopenconns)

	db.SingularTable(true)
	return db
}

// 关闭一个db
func api_close(db *gorm.DB) {
	db.Close()
}

// 创建一张表
func api_create_table(db *gorm.DB, model interface{}) error {
	/*if db.HasTable(model) {
		return nil
	}*/
	//return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(model).Error
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(model).Error
}

// 指定表名创建表
func api_create_table2(db *gorm.DB, tableName string, model interface{}) error {
	if tableName == "" { //使用model表名
		if !db.HasTable(model) {
			return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").CreateTable(model).Error
		}
	} else {
		if !db.HasTable(tableName) {
			return db.Table(tableName).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").CreateTable(model).Error
		}
	}
	return nil
}

func GetInst() *gorm.DB {
	if game_inst == nil {
		connectInst()
	}
	return game_inst
}

func GetLog() *gorm.DB {
	if game_log == nil {
		game_log = connectLog()
	}
	return game_log
}
func GetWeb() *gorm.DB {
	if game_web == nil {
		game_web = connectWeb()
	}
	return game_web
}

func GetLog2() *gorm.DB {
	return connectLog()
}

//// 注册model
func RegisteModel(model interface{}) {
	err := api_create_table(GetInst(), model)
	if err != nil {
		log.Errorf("mysql create table err:%v", err.Error())
	}
}

//// 注册model
func RegisteModel_log(model interface{}) {
	err := api_create_table(GetLog(), model)
	if err != nil {
		log.Errorf("mysql create table err:%v", err.Error())
	}
}

//// 注册model
func RegisteModel_web(model interface{}) {
	err := api_create_table(GetWeb(), model)
	if err != nil {
		log.Errorf("mysql create table err:%v", err.Error())
	}
}

// 注册model,可指定表名
//func RegisteModel2(tableName string, model ibase{}) {
//	for _, db := range db_conns {
//		api_create_table2(db, tableName, model)
//	}
//}
