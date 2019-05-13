package main

import (
	_ "pyg/pyg/routers"
	"github.com/astaxie/beego"
	_ "pyg/pyg/models"
)

func main() {
	beego.Run()
}

