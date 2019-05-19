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
    //生鲜首页
    beego.Router("/index_sx",&controllers.GoodsController{},"get:ShowIndexSx")
    //商品详情
    beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowDetail")
    //同一类型所有商品
    beego.Router("/goodsType",&controllers.GoodsController{},"get:ShowList")
    //商品搜索
    beego.Router("/search",&controllers.GoodsController{},"post:HandleSearch")
    //添加购物车
    beego.Router("/addCart",&controllers.CartController{},"post:HandleAddCart")
    //展示购物车
    beego.Router("/user/showCart",&controllers.CartController{},"get:ShowCart")
    //更改购物车数量-添加
    beego.Router("/upCart",&controllers.CartController{},"post:HandleUpCart")
    //删除购物车商品
    beego.Router("/deleteCart",&controllers.CartController{},"post:HandleDeleteCart")
    //添加商品到订单
    beego.Router("/user/addOrder",&controllers.OrderController{},"post:ShowOrder")
    //提交订单
    beego.Router("/pushOrder",&controllers.OrderController{},"post:HandlePushOrder")
	//展示用户中心订单页
	beego.Router("/user/userOrder",&controllers.UserController{},"get:ShowUserOrder")
	//支付
	beego.Router("/pay",&controllers.OrderController{},"get:Pay")
}

func guolvFunc(ctx*context.Context){
	//过滤校验
	name := ctx.Input.Session("name")
	if name == nil {
		ctx.Redirect(302,"/login")
		return
	}

}
