package models

import (
	"log"
	"gitlab-config-server/config"

	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type BaseModel struct {
	//gorm.Model
}

/**
操作权限
 */
type Operation struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

var DB *gorm.DB

func init() {
	dbConfig := config.GetDbConfig()
	var err error
	DB, err = gorm.Open("mysql",
		dbConfig.User+":"+dbConfig.Passwd+"@tcp("+dbConfig.Host+")/"+dbConfig.Name+"?multiStatements=true&charset=utf8mb4&loc=Asia%2FShanghai")
	if err != nil {
		log.Fatalln("数据库连接创建失败", err.Error())
	}
	if config.ENV != "PROD" {
		DB.LogMode(true)
	}
}
func (m *BaseModel) Error(e error) {
	beego.Error(e.Error())
}
func (m *BaseModel) Info(msg string) {
	beego.Info(msg)
}
