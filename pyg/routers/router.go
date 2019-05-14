package routers

import (
	"pyg/pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	//路由过滤器
	beego.InsertFilter("/user/*",beego.BeforeExec,guolvFunc)
    beego.Router("/", &controllers.MainController{})
    //用户注册
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    //发送短信
    beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
    //邮箱激活
    beego.Router("/register-email",&controllers.UserController{},"get:ShowEmail;post:HandleEmail")
    //激活用户
    beego.Router("/active",&controllers.UserController{},"get:Active")
    //登录实现
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
    //首页实现
    beego.Router("/index",&controllers.GoodsController{},"get:ShowIndex")
    //退出登录
    beego.Router("/user/logout",&controllers.UserController{},"get:Logout")
    //展示用户中心页
    beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")
    //收货地址页
    beego.Router("/user/site",&controllers.UserController{},"get:ShowSite;post:HandleSite")
}

func guolvFunc(ctx*context.Context){
	//过滤校验
	name := ctx.Input.Session("name")
	if name == nil {
		ctx.Redirect(302,"/login")
		return
	}

}
