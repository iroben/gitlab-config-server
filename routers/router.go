package routers

import (
	"gitlab-config-server/controllers"
	"gitlab-config-server/filters"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.Router("/*", &controllers.MainController{}, "options:Options")
	beego.InsertFilter("/v1/*", beego.BeforeRouter, func(ctx *context.Context) {
		filters.Check(ctx)
	})
	beego.Router("/gitlab/login", &controllers.GitLabController{}, "get:Login")
	beego.Router("/gitlab/callback", &controllers.GitLabController{}, "get:Callback")

	ns := beego.NewNamespace("/v1",
		beego.NSRouter("/generate", &controllers.GenerateConfig{}, "post:Generate"),
		beego.NSRouter("/config/yml", &controllers.Config{}, "post:Yml"),
		beego.NSRouter("/config", &controllers.Config{}),
		beego.NSRouter("/log", &controllers.ActiveLog{}),
		beego.NSRouter("/userinfo", &controllers.MainController{},"get:UserInfo"),
		beego.NSRouter("/project", &controllers.GitLabController{}, "get:Projects"),

	)
	beego.AddNamespace(ns)
}
