package models

import "github.com/astaxie/beego/orm"
import (
	_ "github.com/go-sql-driver/mysql"
	"time"
)
/*
orm设置对象关系
1:1 rel(one),reverse(one)
1:n rel(fk),reverse(many)
n:n rel(m2m),reverse(many)
*/
//定义用户类
type User struct {
	Id int
	Name string
	Pwd string
	Articles []*Article `orm:"reverse(many)"`
}

//定义文章类
type Article struct {
	Id int 			`orm:"pk;auto"`
	Title string    `orm:"unique;size(40)"`
	Content string  `orm:"size(500)"`
	Img string		`orm:"null"`
	Time time.Time	`orm:"type(datetime);auto_now_add"`
	ReadCount int	`orm:"default(0)"`
	ArticleType *ArticleType `orm:"rel(fk);null;on_delete(set_null)"`
	Users []*User   `orm:"rel(m2m)"`
}

//定义文章类型类
type ArticleType struct {
	Id int
	TypeName string		`orm:"unique"`
	Articles []*Article  `orm:"reverse(many)"`
}

//orm生成表
func init()  {
	//注册数据库
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/mydb1")
	//注册表
	orm.RegisterModel(new(User),new(Article),new(ArticleType))
	//运行生成表
	orm.RunSyncdb("default",false,true)
}
