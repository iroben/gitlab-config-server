package filters

import (
	"github.com/astaxie/beego/context"
	"gitlab-config-server/services"
)

var (
	UnCheckUrl = []string{
		"POST:/v1/generate",
	}
)

func Check(ctx *context.Context) {
	method := ctx.Input.Method()
	path := ctx.Input.URL()
	for _, v := range UnCheckUrl {
		if v == method+":"+path {
			return
		}
	}
	token := ctx.Input.Query("token")
	result, err := services.Redis.Exists("sid:" + token).Result()
	if err != nil || result == 0 {
		ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.ResponseWriter.WriteHeader(401)
		return
	}

}
