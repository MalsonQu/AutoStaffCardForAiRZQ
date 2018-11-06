package Model

import (
	. "autoStaffCardForAiRZQ/Config"
	"bytes"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
)

func init() {
	// 注册驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)

	var buffer bytes.Buffer

	buffer.WriteString(Config.Db.UserName)
	buffer.WriteString(`:`)
	buffer.WriteString(Config.Db.Password)
	buffer.WriteString(`@tcp(`)
	buffer.WriteString(Config.Db.Host)
	buffer.WriteString(`:`)
	buffer.WriteString(Config.Db.Port)
	buffer.WriteString(`)/`)
	buffer.WriteString(Config.Db.TableName)

	_params := url.Values{}
	_params.Add(`charset`, Config.Db.Charset)
	_params.Add(`loc`, Config.Db.Location)

	_dataBaseUrl := buffer.String() + `?` + _params.Encode()

	// 注册数据库
	orm.RegisterDataBase("default", "mysql", _dataBaseUrl, 30)

	// 定义数据表
	orm.RegisterModel(
		new(User),
		new(StaffLog),
	)

	orm.RunSyncdb("default", false, false)
}
