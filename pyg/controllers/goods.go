package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
)

type GoodsController struct {
	beego.Controller
}

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
