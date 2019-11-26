package services

import (
	"gitlab-config-server/config"

	"github.com/go-redis/redis"
	"log"
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
		log.Println("ERROR: redis init fail: ", err.Error())
	}
	log.Println("redis init: ", pong)
}
