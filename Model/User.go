package Model

import (
	"github.com/astaxie/beego/orm"
)

type User struct {
	Id       uint64      `json:"id"`           // 用户Id
	Name     string      `orm:"size(255)"`     // 用户姓名
	Email    string      `orm:"size(255)"`     // 用户邮箱
	StaffLog []*StaffLog `orm:"reverse(many)"` // 用户的日志
	WeChatId string      `orm:"size(255)"`     // 微信Id
	Trashed  bool        `orm:"index"`         // 用户是否删除
}

func (m *User) GetAll() (*[]*User, error) {
	var result []*User

	o := orm.NewOrm()
	_, err := o.QueryTable(m).Filter("Trashed", false).All(&result)

	return &result, err
}
