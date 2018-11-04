package Model

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 注册驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// 注册数据库
	//orm.RegisterDataBase("default", "mysql", "root:913522@tcp(127.0.0.1:3306)/auto_staff_card?charset=utf8mb4&loc=Asia%2FShanghai", 30)
	orm.RegisterDataBase("default", "mysql", "auto_staff_card:4ChkPSNB58HjPfDf@tcp(127.0.0.1:3306)/auto_staff_card?charset=utf8mb4&loc=Asia%2FShanghai", 30)

	// 定义数据表
	orm.RegisterModel(
		new(User),
		new(StaffLog),
	)

	orm.RunSyncdb("default", false, false)
}
