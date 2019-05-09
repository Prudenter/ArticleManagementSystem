package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"github.com/astaxie/beego/orm"
	"FMSProject/models"
	"encoding/base64"
)

type UserController struct {
	beego.Controller
}

//注册展示页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//注册业务处理
func (this *UserController) HandleRegister() {
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		fmt.Println("用户名和密码不能为空！")
		this.TplName = "register.html"
		return
	}
	//处理数据
	o:=orm.NewOrm()
	var user models.User
	user.Name = userName
	user.Pwd = pwd
	_,err := o.Insert(&user)
	if err != nil {
		fmt.Println("注册失败，请重新注册！")
		this.TplName = "register.html"
		return
	}
	//渲染
	//this.TplName = "login.html"
	//重定向
	this.Redirect("/login",302)
}

//登录展示页面
func (this *UserController) ShowLogin() {
	//获取coolie数据，如果有值，说明上一次记住了用户名，默认勾选记住用户名框，否则不记住用户名，不勾选记住用户名框
	userName := this.Ctx.GetCookie("userName")
	//解密，获取中文登录名
	dec,_:= base64.StdEncoding.DecodeString(userName)
	if userName != ""{
		this.Data["userName"] = string(dec)
		this.Data["checked"] = "checked"
	}else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}
	this.TplName = "login.html"
}

//登录业务处理
func (this *UserController) HandleLogin() {
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		fmt.Println("用户名和密码不能为空！")
		this.TplName = "login.html"
		return
	}
	//处理数据-查询校验
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		fmt.Println("用户名不存在！")
		this.TplName = "login.html"
		return
	}
	//校验密码
	if user.Pwd != pwd {
		fmt.Println("密码错误，请重新输入！")
		this.TplName = "login.html"
		return
	}
	//实现记住用户名功能  上一次登陆成功以后，点击了记住用户名，下一次登陆的时候默认显示用户名
	//1.登陆成功后，设置cookie，记住用户名
	//2.判断是否勾选了记住用户名
	remember := this.GetString("remember")
	//给userName加密，解决cookie不支持中文问题
	enc := base64.StdEncoding.EncodeToString([]byte(userName))
	if remember == "on" {
		// key value  存活时间
		this.Ctx.SetCookie("userName",enc,5000)
	}else {
		//设置存活时间为-1，使cookie失效
		this.Ctx.SetCookie("userName",userName,-1)
	}

	//设置session,用于页面显示
	this.SetSession("userName",userName)

	//跳转到首页
	this.Redirect("/article/index",302)
}

//用户退出业务处理
func (this *UserController) HandleLogout() {
	//删除session，然后跳转到登录页面
	this.DelSession("userName")
	this.Redirect("/login",302)
}