package routers

import (
	"github.com/astaxie/beego"
	"pygproject/pyg/controllers"
)

func init() {
	//----------------------------------用户模块-----------------------------------------------------------------------
	//用户注册
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister")



}
