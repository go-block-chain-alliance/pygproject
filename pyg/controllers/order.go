package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"github.com/smartwalle/alipay"
)

type OrderController struct {
	beego.Controller
}

//展示订单页面
func(this*OrderController)ShowOrder(){
	//获取数据
	goodsIds := this.GetStrings("checkGoods")

	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart",302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	name := this.GetSession("name")

	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",name.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	conn,_ := redis.Dial("tcp","192.168.179.65:6379")

	//获取商品,获取总价和总件数
	var goods []map[string]interface{}
	var totalPrice ,totalCount int

	for _,v := range goodsIds{
		temp := make(map[string]interface{})
		id,_ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		//获取商品数量
		count,_ := redis.Int(conn.Do("hget","cart_"+name.(string),id))

		//计算小计
		littlePrice := count * goodsSku.Price


		//把商品信息放到行容器
		temp["goodsSku"] = goodsSku
		temp["count"] = count
		temp["littlePrice"] = littlePrice

		totalPrice += littlePrice
		totalCount += 1

		goods = append(goods,temp)

	}

	//返回数据
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goods"] = goods
	this.Data["goodsIds"] = goodsIds
	this.TplName = "place_order.html"
}

//提交订单
func(this*OrderController)HandlePushOrder(){
	//获取数据
	addrId,err1 := this.GetInt("addrId")
	payId,err2 := this.GetInt("payId")
	goodsIds := this.GetString("goodsIds")
	totalCount,err3 := this.GetInt("totalCount")
	totalPrice ,err4 := this.GetInt("totalPrice")



	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)

	name := this.GetSession("name")
	if name == nil{
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录"
		return
	}

	//校验数据
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || goodsIds == ""{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整"
		return
	}
	//处理数据
	//把数据插入到mysql数据库中
	//获取用户对象和地址对象
	o := orm.NewOrm()
	var user models.User
	user.Name = name.(string)
	o.Read(&user,"Name")

	var address models.Address
	address.Id = addrId
	o.Read(&address)

	var orderInfo models.OrderInfo

	orderInfo.User = &user
	orderInfo.Address = &address
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = 10
	orderInfo.OrderId = time.Now().Format("20060102150405"+strconv.Itoa(user.Id))
	//开启事务
	o.Begin()
	o.Insert(&orderInfo)

	conn,_:=redis.Dial("tcp","192.168.179.65:6379")

	defer conn.Close()
	//插入订单商品
	//goodsIds  //2  3  5
	goodsSlice:= strings.Split(goodsIds[1:len(goodsIds)-1]," ")
	for _,v := range goodsSlice{
		//插入订单商品表

		//获取商品信息
		id,_ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		oldStock := goodsSku.Stock
		beego.Info("原始库存等于",oldStock)

		//获取商品数量
		count,_ := redis.Int(conn.Do("hget","cart_"+name.(string),id))

		//获取小计
		littlePrice := goodsSku.Price * count

		//插入
		var orderGoods models.OrderGoods
		orderGoods.OrderInfo = &orderInfo
		orderGoods.GoodsSKU = &goodsSku
		orderGoods.Count = count
		orderGoods.Price = littlePrice
		//插入之前需要更新商品库存和销量
		if goodsSku.Stock < count{
			resp["errno"] = 4
			resp["errmsg"] = "库存不足"
			o.Rollback()
			return
		}
		//goodsSku.Stock -= count
		//goodsSku.Sales += count

		o.Read(&goodsSku)

		qs := o.QueryTable("GoodsSKU").Filter("Id",id).Filter("Stock",oldStock)
		num,_:= qs.Update(orm.Params{"Stock":goodsSku.Stock - count,"Sales":goodsSku.Sales+count})
		if num == 0 {
			resp["errno"] = 7
			resp["errmsg"] = "购买失败，请重新排队！"
			o.Rollback()
			return
		}



		_,err := o.Insert(&orderGoods)
		if err != nil {
			resp["errno"] = 3
			resp["errmsg"] = "服务器异常"
			o.Rollback()
			return
		}
		_,err = conn.Do("hdel","cart_"+name.(string),id)
		if err != nil {
			resp["errno"] = 6
			resp["errmsg"] = "清空购物车失败"
			o.Rollback()
			return
		}

	}


	//返回数据
	o.Commit()
	resp["errno"] = 5
	resp["errmsg"] = "OK"
}

//支付
func(this*OrderController)Pay(){
	//获取数据
	orderId,err:=this.GetInt("orderId")
	if err != nil {
		this.Redirect("/user/userOrder",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	orderInfo.Id = orderId
	o.Read(&orderInfo)

	//支付


	//appId, aliPublicKey, privateKey string, isProduction bool
	publiKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwG0RShI/lJzJNaZOZxZT
c963DaEJyL8brVNpbxZl4BzWjJ60iIdFFl6zyoThJtPedH2S+wI9OgKMVPpnOrgR
KYv5Tl+xB8F6KlP1dkFIDDLR+CodNCoyElL85bFnSgYkj7cSH/Ve4aHXf+wo1EYk
Gz1urludvU8HMDdXWJsogcKrdgQgrJJeiVkKdmsT49Xoc+AdYSGpA7Av0TsikGns
HeIJITTz7MAYaVLVw4sGddEJDyZWpyL5G7KqHJogS85fTa+WK6NCKmCa7NiqNZfp
wCy0jEb4qT1ip+dZYiCiNM4/JahCMh38coNRszdm0dpNoVwucwFlxVpZ2reGNU1r
3QIDAQAB`

	privateKey := `MIIEpQIBAAKCAQEAwG0RShI/lJzJNaZOZxZTc963DaEJyL8brVNpbxZl4BzWjJ60
iIdFFl6zyoThJtPedH2S+wI9OgKMVPpnOrgRKYv5Tl+xB8F6KlP1dkFIDDLR+Cod
NCoyElL85bFnSgYkj7cSH/Ve4aHXf+wo1EYkGz1urludvU8HMDdXWJsogcKrdgQg
rJJeiVkKdmsT49Xoc+AdYSGpA7Av0TsikGnsHeIJITTz7MAYaVLVw4sGddEJDyZW
pyL5G7KqHJogS85fTa+WK6NCKmCa7NiqNZfpwCy0jEb4qT1ip+dZYiCiNM4/JahC
Mh38coNRszdm0dpNoVwucwFlxVpZ2reGNU1r3QIDAQABAoIBAQCsWIOlvhZoOs0U
WjHarupr20xEzrl+rXxSj2TddEgmpG2dYP/9UHqWgJezibRHHHggCeC9JNJFxMZ/
zg7rTrVAavgONDLQ6X9LrgspsWqgUlwxUzb449oZA28zIuOKL1pLxgJb0V775AKp
tpETHwdzxl/9llz/k2qyyr5WxBFRtcSrmLWEUZ5G2HMtHvkKxRt01FRbP11oIbmZ
KbNJDq3CTC0H8dyIphWvAN89e2JdTYBOyDUD39AFV+SzVqnu+8mbEv+XYg1bCP4J
aaqbAr1JROgzEVssvfOE52bmCLdmE6JavkMqaC+uA+hFqVdWE2nLFPWaWz3NL+5d
dD648ZZBAoGBAPNdObtQC0varL5vgvCyQ0k4qs42tZ3cWUglP2ttVLnTR7BBr6x6
yECcdtO1ZwMRiTxEsz2oWaneBTu2hPAr0IdEreZsFokTDb9LITKCtkB1TMceFDWE
1cZvI6FwfjxVk1JeNk7yKm5To2oqN8rxMv52FYUzVz64Zz8EVKv/SIhJAoGBAMpq
xvRSEq8+FA+kZtLyW/uIxUnLX3ivjAAtsQw/vtbJdMgjZWeRf2iMGa0EVW+uEwnf
wbcnQXSNHVwRMnJiK9gpig/KkWLqaM7eZrRubeS6dBwLUhfEwaBg8hJE3gLA429q
gML3m93Iij0HAc16FMChclvCjXIbV9XlhqP39w71AoGBAJK7EMXpOwZfMGwZm983
++235vQydEpbwtEG9Df3UXBA/SY+VIcv+HFMZTC8XQGynwXhfhic2oLaxFj+cSTF
phMIy7j0TpoTDOTbjYaA3RX8I3CiqBikoKfl9put0c7a4dp1x1TOGdsvPoYSMlWA
G/jkhZEsJVxBnq6WE98oKjlRAoGBAJhK1f2kcmJe9oD+VE6KAiKxuJ3Y4a/PhCnu
NrLckxzO3Ypm9ziBA7cJEZhXFmC8O57GNt0yL9EdCuXmGmps6kfsmO9gnRoq+0gJ
lIRUQWJB1nHzIoS3iGa+CeMs5Ux1C6kcHFyUJzUqWLepufV60HpN/diD/B/J6sAH
vNFJExyxAoGABD2WiL20YVuPFl7S29Z/jFfzPHaz7ws+AqPSBCw+O9r7qKsvcc8D
9QHS9mHC/sGKY/LPB/x4RZiYTRG+Fk1GOg/Awtc82hRgM4CragQN3KX2cMO3p0vT
OZLHuX1KVt2gcnRVMYVPYSVR1/pRZFONfzoRL04lUttpDOWuWrDxiuc=`
	client := alipay.New("2016092200569649",publiKey,privateKey,false)
	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://192.168.179.65:8080/payOK"
	p.ReturnURL = "http://192.168.179.65:8080/payOK"
	p.Subject = "品优购"
	p.OutTradeNo = orderInfo.OrderId
	p.TotalAmount = strconv.Itoa(orderInfo.TotalPrice)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		beego.Error("支付失败")
	}
	payUrl := url.String()
	this.Redirect(payUrl,302)
}