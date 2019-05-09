package routers

import (
	"FMSProject/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)
func init() {

	//建立路由过滤器,由于登录校验   第一个参数是过滤匹配支持正则    过滤位置      过滤操作（函数） 参数是context
	beego.InsertFilter("/article/*",beego.BeforeExec,filterFunc)

    beego.Router("/", &controllers.MainController{})
	//注册业务处理
	beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
	//登录业务处理
	beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
	//首页展示
	beego.Router("/article/index",&controllers.ArticleController{},"get,post:ShowIndex")
	//添加文章
	beego.Router("/article/addArticle",&controllers.ArticleController{},"get:ShowAddArticle;post:HandleAddArticle")
	//查看文章详情
	beego.Router("/article/content",&controllers.ArticleController{},"get:ShowContent")
	//编辑文章
	beego.Router("/article/update",&controllers.ArticleController{},"get:ShowUpdate;post:HandleUpdate")
	//删除文章
	beego.Router("/article/delete",&controllers.ArticleController{},"get:HandleDelete")
	//添加文章类型
	beego.Router("/article/addType",&controllers.ArticleController{},"get:ShowAddType;post:HandleAddType")
	//删除文章类型
	beego.Router("/article/deleteType",&controllers.ArticleController{},"get:HandleDeleteType")
	//退出登录
	beego.Router("/article/logout",&controllers.UserController{},"get:HandleLogout")
}

func filterFunc(ctx *context.Context)  {
	//登录校验
	//设置session
	//ctx.Output.Session("userName","ddd")
	//获取session
	userName := ctx.Input.Session("userName")
	if userName == nil {
		//跳转到登录页面
		ctx.Redirect(302,"/login")
		return
	}
}