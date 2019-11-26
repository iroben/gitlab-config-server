package main

import (
	_ "gitlab-config-server/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

