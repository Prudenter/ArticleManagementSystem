package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"path"
	"time"
	"github.com/astaxie/beego/orm"
	"FMSProject/models"
	"math"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"bytes"
	"encoding/gob"
)

//定义controller
type ArticleController struct {
	beego.Controller
}

//定义函数，获取session中的数据,用于页面显示
func (this *ArticleController) GetSessionUserName() string {
	//获取session,用于页面显示
	userName := this.GetSession("userName")
	return userName.(string) //使用类型断言，将interface转换为string
}

//首页展示
func (this *ArticleController) ShowIndex() {

	//调用方法，获取当前登录的用户名
	this.Data["userName"] = this.GetSessionUserName();

	//获取所有文章数据展示到页面上
	o := orm.NewOrm()
	qs := o.QueryTable("Article")
	//获取选中的文章类型
	typeName := this.GetString("select")
	//定义文章结构体切片
	var articles []models.Article
	var count int64
	if typeName == "" {
		//获取总记录数
		count, _ = qs.RelatedSel("ArticleType").Count()
	} else {
		count, _ = qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).Count()
	}
	//定义每页的记录数
	pageCount := 2
	//计算总页数
	pageNum := math.Ceil(float64(count) / float64(pageCount))
	//获取首页末页记录
	//获取页码数
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1 //默认跳转到首页，跳转时自动赋值
	}

	if typeName == "" {
		//获取对应页码的所有数据
		qs.Limit(pageCount, pageCount*(pageIndex-1)).RelatedSel("ArticleType").All(&articles)
	} else {
		//获取对应页码的本类型数据
		qs.Limit(pageCount, pageCount*(pageIndex-1)).RelatedSel("ArticleType").
			Filter("ArticleType__TypeName", typeName).All(&articles)
	}
	//查询所有文章类型并返回到页面
	var articleTypes []models.ArticleType

	//连接redis数据库,从redis数据库中查询数据
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("redis数据库连接失败！")
		return
	}
	defer conn.Close()
	//从redis数据库中查询所有文章类型
	resp, err := conn.Do("get", "articleTypes")
	if err != nil {
		fmt.Println("数据获取失败！")
		return
	}
	//判断数据是否为空，如果为空说明是第一次登录，需要从mysql数据库中获取数据,并将数据存入redis数据库中
	ret, _ := redis.Bytes(resp, err)
	if len(ret) == 0 {
		o.QueryTable("ArticleType").All(&articleTypes)
		//将数据存入redis数据库中,因为是结构体切片，所以需要序列化存入-字节化
		var buffer bytes.Buffer
		//定义一个编码器
		enc := gob.NewEncoder(&buffer) //buffer的作用是什么？
		//将数据编码
		enc.Encode(articleTypes)
		//将字节化的数据存入redis数据库中
		conn.Do("set", "articleTypes", buffer.Bytes())
		fmt.Println("从mysql中获取数据")
	} else {
		//如果获取到的数据不为空,则需要进行解码操作
		//定义一个解码器
		dec := gob.NewDecoder(bytes.NewReader(ret))  //NewReader的作用是什么?这一步时已经解码了？
		//接受解码后的数据
		dec.Decode(&articleTypes)
		fmt.Println("从redis中获取数据")
		fmt.Println(articleTypes)
	}

	this.Data["articleTypes"] = articleTypes
	//返回选中的文章类型，用于前端页面比较，显示下拉框
	this.Data["typeName"] = typeName
	//返回数据到页面
	this.Data["articles"] = articles
	this.Data["count"] = count
	this.Data["pageNum"] = pageNum
	this.Data["pageIndex"] = pageIndex
	//视图布局：把大框（模板）和主要部分拼接
	this.Layout = "layout.html"
	//使用LayoutSections将js代码从模板中分离出来，只用于index页面
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["indexJs"] = "indexJs.html"
	this.TplName = "index.html"
}

//添加文章页面展示
func (this *ArticleController) ShowAddArticle() {

	//调用方法，获取当前登录的用户名
	this.Data["userName"] = this.GetSessionUserName();

	//获取所有文章类型返回到前台页面
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes
	//把大框和主要部分拼接
	this.Layout = "layout.html"
	this.TplName = "add.html"
}

//添加文章业务处理
func (this *ArticleController) HandleAddArticle() {
	//获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	//调用函数处理文件上传-包含文件获取，校验，存储，返回文件存储路径
	savePath := UploadFile(this, "uploadname")
	//获取文章类型名
	typeName := this.GetString("select")
	//校验数据
	if articleName == "" || content == "" || typeName == "" {
		fmt.Println("获取数据为空或文件上传错误！")
		this.Data["errmsg"] = "获取数据为空或文件上传错误！"
		//把大框和主要部分拼接
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var art models.Article
	art.Title = articleName
	art.Content = content
	art.Img = savePath
	//***获取一个文章类型对象，并插入到文章中
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Read(&articleType, "TypeName")
	art.ArticleType = &articleType

	//插入数据
	_, err := o.Insert(&art)
	if err != nil {
		fmt.Println("添加文件失败！请重新添加！")
		this.Data["errmsg"] = "添加文件失败，请重新添加！"
		//把大框和主要部分拼接
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	//返回数据，跳转页面
	this.Redirect("/article/index", 302)
}

//查看文章详情业务处理
func (this *ArticleController) ShowContent() {

	//调用方法，获取当前登录的用户名
	this.Data["userName"] = this.GetSessionUserName();

	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("文章不存在，请重新查询！")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据
	//查询
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Read(&article)

	//联合查询，查询articleType中的typename,用于查看详情时类型展示
	//beego默认执行的是级联删除——这里我们设置的级联删除模式为set_null,删除文章类型时，不删除文章，
	// 因此查询时需要判断article.ArticleType为null的情况
	if article.ArticleType != nil {
		qs := o.QueryTable("Article")
		qs.RelatedSel("ArticleType").Filter("ArticleType__Id", article.ArticleType.Id).All(&article)
	}

	//每查询一次，修改阅读量
	article.ReadCount += 1
	o.Update(&article)

	//插入多对多关系，增加浏览记录
	//根据用户名获取用户对象
	var user models.User
	user.Name = this.GetSessionUserName()
	o.Read(&user, "Name")
	/*
	多对多插入操作：
	1.获取orm对象
	2.获取被插入数据的对象，文章
	3.获取需要插入的对象，用户
	4.获取多对多操作对象
	5.多对多对象插入数据-本质上数据被插入到多对多关系表中
	*/
	//获取多对多操作对象
	m2m := o.QueryM2M(&article, "Users")
	//用多对多操作对象插入数据
	m2m.Add(user)
	//多对多查询一，LoadRelated,不能去重
	//o.LoadRelated(&article,"Users")
	//多对多查询二，使用高级查询，可以去重
	var users []models.User
	//需要哪个就查哪个表
	o.QueryTable("User").Filter("Articles__Article__Id", id).Distinct().All(&users)
	//返回用户集合到前台
	this.Data["users"] = users
	this.Data["article"] = article
	//把大框和主要部分拼接
	this.Layout = "layout.html"
	this.TplName = "content.html"
}

//编辑文章页面展示
func (this *ArticleController) ShowUpdate() {

	//调用方法，获取当前登录的用户名
	this.Data["userName"] = this.GetSessionUserName();

	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("文章不存在，请重新编辑！")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据
	//查询
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Read(&article)
	//返回数据到前台页面
	this.Data["article"] = article
	//把大框和主要部分拼接
	this.Layout = "layout.html"
	this.TplName = "update.html"
}

//封装处理文件上传的函数
func UploadFile(this *ArticleController, formFileName string) string {
	//获取图片
	file, head, err := this.GetFile(formFileName)
	if err != nil {
		fmt.Println("图片上传失败！")
		return ""
	}
	defer file.Close()
	//校验文件大小
	if head.Size > 5000000 {
		fmt.Println("图片数据过大，请重新上传！")
		return ""
	}
	//校验文件格式
	//获取文件后缀名
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		fmt.Println("图片格式不正确，请重新上传！")
		return ""
	}
	//防止重名
	fileName := time.Now().Format("200601021504057777")
	//把上传的文件存储到目标文件夹
	this.SaveToFile(formFileName, "./static/img/"+fileName+ext)
	return "/static/img/" + fileName + ext
}

//编辑文章业务处理
func (this *ArticleController) HandleUpdate() {
	//获取数据
	id, _ := this.GetInt("id") //通过隐藏域传值
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	savePath := UploadFile(this, "uploadname")
	//校验数据
	if articleName == "" || content == "" {
		fmt.Println("获取数据为空或文件上传错误！")
		this.Data["errmsg"] = "获取数据为空或文件上传错误！"
		this.Redirect("/article/update?id="+strconv.Itoa(id), 302)
		return
	}
	//处理数据
	//更新数据
	o := orm.NewOrm()
	var article models.Article
	//先查询要更新的数据是否存在
	article.Id = id
	err := o.Read(&article)
	if err != nil {
		fmt.Println("更新数据不存在，请重新操作！")
		this.Data["errmsg"] = "更新数据不存在，请重新操作！"
		this.Redirect("/article/update?id="+strconv.Itoa(id), 302)
		return
	}
	//更新
	article.Title = articleName
	article.Content = content
	article.Img = savePath
	o.Update(&article)
	//返回数据
	this.Redirect("/article/index", 302)
}

//删除文章业务处理
func (this *ArticleController) HandleDelete() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("文章不存在，请重新删除！")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Delete(&article, "Id")
	//返回首页
	this.Redirect("/article/index", 302)
}

//添加文章类型页面展示
func (this *ArticleController) ShowAddType() {

	//调用方法，获取当前登录的用户名
	this.Data["userName"] = this.GetSessionUserName();

	//查询所有文章类型显示到页面上
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").OrderBy("Id").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes
	//把大框和主要部分拼接
	this.Layout = "layout.html"
	//使用LayoutSections将js代码从模板中分离出来，只用于addType页面
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["indexJs"] = "addTypeJs.html"
	this.TplName = "addType.html"
}

//添加文章类型页面展示
func (this *ArticleController) HandleAddType() {
	//获取数据
	typeName := this.GetString("typeName")
	//校验数据
	if typeName == "" {
		fmt.Println("文章类型不能为空！")
		//把大框和主要部分拼接
		this.Layout = "layout.html"
		this.TplName = "addType.html"
		return
	}
	//处理数据
	//添加类型
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Insert(&articleType)
	//返回数据
	this.Redirect("/article/addType", 302)
}

//删除文章类型业务处理
func (this *ArticleController) HandleDeleteType() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("文章类型不存在，请重新删除！")
		this.Redirect("/article/addType", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id = id
	//beego默认执行的是级联删除，cascade——删除文章类型时，会顺便删掉所有属于该类型的文章
	// 这里我们设置的删除模式是set_null,删除文章类型时，不删除文章
	o.Delete(&articleType, "Id")
	//返回首页
	this.Redirect("/article/addType", 302)
}
