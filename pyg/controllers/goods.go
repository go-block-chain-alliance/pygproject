package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"math"
	"github.com/gomodule/redigo/redis"
)

type GoodsController struct {
	beego.Controller
}

//展示首页
func(this*GoodsController)ShowIndex(){
	name := this.GetSession("name")
	if name != nil {
		this.Data["name"] = name.(string)
	}else {
		this.Data["name"] = ""
	}

	//获取类型信息并传递给前段
	//获取一级菜单
	o := orm.NewOrm()
	//接受对象
	var oneClass []models.TpshopCategory
	//查询
	o.QueryTable("TpshopCategory").Filter("Pid",0).All(&oneClass)


	//获取第二级
	var types []map[string]interface{}//定义总容器
	for _,v := range oneClass{
		//行容器
		t := make(map[string]interface{})

		var secondClass []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid",v.Id).All(&secondClass)
		t["t1"] = v  //一级菜单对象
		t["t2"] = secondClass  //二级菜单集合
		//把行容器加载到总容器中
		types = append(types,t)
	}

	//获取第三季菜单
	for _,v1 := range types{
		//循环获取二级菜单
		var erji []map[string]interface{} //定义二级容器
		for _,v2 := range v1["t2"].([]models.TpshopCategory){
			t := make(map[string]interface{})
			var thirdClass []models.TpshopCategory
			//获取三级菜单
			o.QueryTable("TpshopCategory").Filter("Pid",v2.Id).All(&thirdClass)
			t["t22"] = v2  //二级菜单
			t["t23"] = thirdClass   //三级菜单
			erji = append(erji,t)
		}
		//把二级容器放到总容器中
		v1["t3"] = erji
	}


	this.Data["types"] = types
	this.TplName = "index.html"
}

//展示生鲜首页
func(this*GoodsController)ShowIndexSx(){
	//获取生鲜首页内容
	//获取商品类型
	o := orm.NewOrm()
	//获取所有类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes
	//获取轮播图
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanners)
	this.Data["goodsBanners"] = goodsBanners

	//获取促销商品
	var promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanners)
	this.Data["promotions"] = promotionBanners

	//获取首页商品展示
	var goods []map[string]interface{}

	for _,v := range goodsTypes{
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		qs:=o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Id",v.Id).OrderBy("Index")
		//获取文字商品
		qs.Filter("DisplayType",0).All(&textGoods)
		//获取图片商品
		qs.Filter("DisplayType",1).All(&imageGoods)

		//定义行容器
		temp := make(map[string]interface{})
		temp["goodsType"] = v
		temp["textGoods"] = textGoods
		temp["imageGoods"] = imageGoods

		//把行容器追加到总容器中
		goods = append(goods,temp)
	}
	this.Data["goods"] = goods

	this.TplName = "index_sx.html"
}

//独立于beego框架的

func PageEdit(pageCount int,pageIndex int)[]int{
	//不足五页
	var pages []int
	if pageCount < 5{
		for i:=1;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else if pageIndex <= 3{
		for i:=1;i<=5;i++{
			pages = append(pages,i)
		}
	}else if pageIndex >= pageCount -2{
		for i:=pageCount - 4;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else {
		for i:=pageIndex - 2;i<=pageIndex + 2;i++{
			pages = append(pages,i)
		}
	}

	return pages
}

//商品详情页
func(this*GoodsController)ShowDetail(){
	//获取数据
	id,err := this.GetInt("Id")
	//校验数据
	if err != nil {
		beego.Error("商品链接错误")
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	//根据id获取商品有关数据
	o := orm.NewOrm()
	var goodsSku models.GoodsSKU
	/*goodsSku.Id = id
	o.Read(&goodsSku)*/
	//获取商品详情
	o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goodsSku)

	//获取同一类型的新品推荐
	var newGoods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsSku.GoodsType.Name)
	qs.OrderBy("-Time").Limit(2,0).All(&newGoods)

	//存储浏览记录
	name := this.GetSession("name")
	if name != nil {
		//把历史浏览记录存储在redis中
		conn,err := redis.Dial("tcp","192.168.179.65:6379")
		if err == nil {
			defer conn.Close()
			conn.Do("lrem","history_"+name.(string),0,id)
			_,err = conn.Do("lpush","history_"+name.(string),id)
			beego.Info(err)
		}
	}


	this.Data["newGoods"] = newGoods
	//传递数据
	this.Data["goodsSku"] = goodsSku
	this.TplName = "detail.html"
}

//展示商品列表页
func(this*GoodsController)ShowList(){
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("类型不存在")
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//获取排序方式
	sort := this.GetString("sort")

	//实现分页

	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id)
	//获取总页码
	count,_ := qs.Count()
	pageSize := 1
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//获取当前页码
	pageIndex,err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := PageEdit(pageCount,pageIndex)
	this.Data["pages"] = pages
	//获取上一页，下一页的值
	var prePage,nextPage int
	//设置个范围
	if pageIndex -1 <= 0{
		prePage = 1
	}else {
		prePage = pageIndex - 1
	}


	if pageIndex +1 >= pageCount{
		nextPage = pageCount
	}else {
		nextPage = pageIndex + 1
	}


	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	qs = qs.Limit(pageSize,pageSize*(pageIndex - 1))

	//获取排序
	if sort == ""{
		qs.All(&goods)
	}else if sort == "price"{
		qs.OrderBy("Price").All(&goods)
	}else {
		qs.OrderBy("-Sales").All(&goods)
	}

	this.Data["sort"] = sort



	//返回数据
	this.Data["id"] = id
	this.Data["goods"] = goods
	this.TplName = "list.html"
}

//搜索页面
func(this*GoodsController)HandleSearch(){
	//获取数据
	goodsName := this.GetString("goodsName")
	//校验数据
	if goodsName == ""{
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//模糊查询
	o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).All(&goods)

	//返回数据
	this.Data["goods"] = goods
	this.TplName = "search.html"
}