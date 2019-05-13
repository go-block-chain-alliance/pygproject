package main

import (
	_ "pygproject/pyg/routers"
	"github.com/astaxie/beego"
	_ "pygproject/pyg/models"
)

func main() {
	beego.Run()
}

