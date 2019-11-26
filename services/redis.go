package services

import (
	"gitlab-config-server/config"

	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

var Redis *redis.Client

func init() {

	Redis = redis.NewClient(&redis.Options{
		Addr:     config.GetString("redisHost"),
		Password: config.GetString("redisPasswd"),
		DB:       config.GetInt("redisDB"),
	})
	pong, err := Redis.Ping().Result()
	if err != nil {
		beego.Error("redis init fail: ", err.Error())
	}
	beego.Info("redis init: ", pong)
}
