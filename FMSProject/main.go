package main

import (
	_ "FMSProject/routers"
	"github.com/astaxie/beego"
	_ "FMSProject/models"
)

func main() {
	//给视图函数建立映射
	beego.AddFuncMap("prePage", prePage)
	beego.AddFuncMap("nextPage", nextPage)
	beego.AddFuncMap("autoKey", AutoKey)
	beego.Run()
}

//计算上一页页数
func prePage(pageIndex int) int {
	if pageIndex <= 1 {
		return 1
	}
	return pageIndex - 1
}

//计算下一页页数
func nextPage(pageIndex int, pageNum float64) int {
	if pageIndex >= int(pageNum) {
		return int(pageNum)
	}
	return pageIndex + 1
}

//实现文章类型页面id自增
func AutoKey(key int) int {
	return key + 1
}
