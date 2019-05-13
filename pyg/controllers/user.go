package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"regexp"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"time"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"math/rand"
	"pygproject/pyg/models"
)

//用户控制器
type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//发送错误信息函数
func respFunc(this *UserController, resp map[string]interface{}) {
	//发送数据
	this.Data["json"] = resp
	//定义发送方式,json方式
	this.ServeJSON()
}

//定义接收数据结构体
type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}

//发送短信
func (this *UserController) HandleSendMsg() {
	//获取数据
	phone := this.GetString("phone")
	//定义容器
	resp := make(map[string]interface{})

	//向前端发送信息
	defer respFunc(this, resp)
	//校验数据
	if phone == "" {
		fmt.Println("获取电话号码失败")
		//给容器赋值
		resp["errnum"] = 1
		resp["errmsg"] = "获取电话号码失败"
		return
	}

	//检查电话号码是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == "" {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误"
		return
	}

	//发短信
	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}
	//生成6位数随机数
	//方法二
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	msgnum := fmt.Sprintf("%06d", rnd.Int31n(1000000))

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = "{\"code\":" + msgnum + "}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		fmt.Println("短信发送失败")
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}

	//json数据解析
	//创建一个接受返回数据的结构体
	var message Message
	//将json数据解析为字符切片，存储在容器message中【注意】
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	if message.Message != "OK" {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}
	//数据发送成功
	resp["errno"] = 5
	resp["errmsg"] = "发送成功"

	//传递生成的验证码给前端，做数据校验
	resp["code"] = msgnum
}

//注册操作具体实现
func (this *UserController) HandleRegister() {
	//获取数据
	phone := this.GetString("phone")
	password := this.GetString("password")
	repassword := this.GetString("repassword")
	//校验数据
	//数据完整性校验
	if phone == "" || password == "" || repassword == "" {
		fmt.Println("输入信息不完整,请重新输入！")
		this.Data["errmsg"] = "输入信息不完整,请重新输入！"
		this.TplName = "register.html"
		return
	}
	//两次输入密码是否一致校验
	if password != repassword {
		fmt.Println("两次输入密码不一致")
		this.Data["errmsg"] = "两次输入密码不一致"
		this.TplName = "register.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = password
	user.Phone = phone
	o.Insert(&user)

	//返回数据 //激活页面
	//储存cookie
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	//跳转到激活邮箱页面
	this.Redirect("/register-email", 302)
}
