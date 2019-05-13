package controllers

import (
	"github.com/astaxie/beego"
)

//用户控制器
type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}
