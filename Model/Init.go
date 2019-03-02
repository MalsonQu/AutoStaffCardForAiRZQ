package Model

import (
	"bytes"
	. "github.com/MalsonQu/AutoStaffCardForAiRZQ/Config"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"os"
)

func init() {
	// 注册驱动
	err := orm.RegisterDriver("mysql", orm.DRMySQL)

	if err != nil {
		os.Exit(0)
		return
	}

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
	err = orm.RegisterDataBase("default", "mysql", _dataBaseUrl, 30)

	if err != nil {
		os.Exit(0)
		return
	}

	// 定义数据表
	orm.RegisterModel(
		new(User),
		new(StaffLog),
	)

	err = orm.RunSyncdb("default", false, false)

	if err != nil {
		os.Exit(0)
		return
	}

}
