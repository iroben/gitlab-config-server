package config

import (
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
)

var (
	ENV       = "DEV"
	appConfig config.Configer
)

type DbConfig struct {
	Name   string
	Host   string
	User   string
	Passwd string
}

func init() {
	var err error
	appConfig, err = config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		beego.Error("load conf/app.conf failed: " + err.Error())
		return
	}
	env := os.Getenv("ENV")
	if len(env) != 0 {
		ENV = env
	}
}

func GetString(key string) (retVal string) {
	retVal = appConfig.String(ENV + "::" + key)
	if len(retVal) > 0 {
		return
	}
	retVal = appConfig.String(key)
	return
}

func GetInt(key string) (retVal int) {
	retVal, _ = appConfig.Int(ENV + "::" + key)
	if retVal > 0 {
		return
	}
	retVal, _ = appConfig.Int(key)
	return
}

func GetBool(key string) (retVal bool) {
	retVal, err := appConfig.Bool(ENV + "::" + key)
	if err == nil {
		return
	}
	retVal, _ = appConfig.Bool(key)
	return
}

func GetDbConfig() *DbConfig {
	return &DbConfig{
		Name:   GetString("dbname"),
		Host:   GetString("dbhost"),
		User:   GetString("dbuser"),
		Passwd: GetString("dbpasswd"),
	}
}
