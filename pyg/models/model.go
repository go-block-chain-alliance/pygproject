package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id int
	Name string `orm:"size(40);unique"`
	Pwd string `orm:"size(40)"`
	Phone string `orm:"size(11)"`
	Email string `orm:"null"`
	Active bool `orm:"default(false)"`
	Addresses []*Address `orm:"reverse(many)"`
}

type Address struct {
	Id int
	Receiver string `orm:"size(40)"`
	Addr string `orm:"size(100)"`
	PostCode string
	Phone string `orm:"size(11)"`
	IsDefault bool `orm:"default(false)"` //默认地址
	User *User `orm:"rel(fk)"`

}

type TpshopCategory struct {
	Id int
	CateName string `orm:"default('')"`
	Pid int `orm:"default(0)"`
	IsShow int `orm:"default(1)"`
	CreateTime int `orm:"null"`
	UpdateTime int `orm:"null"`
	DeleteTime int `orm:"null"`
}

func init(){
	//注册数据库
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/pyg")
	//注册表结构
	orm.RegisterModel(new(User),new(Address),new(TpshopCategory))
	//跑起来
	orm.RunSyncdb("default",false,true)
}