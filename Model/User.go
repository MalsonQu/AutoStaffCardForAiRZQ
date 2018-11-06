package Model

import (
	"github.com/astaxie/beego/orm"
)

type User struct {
	Id       uint64      `json:"id"`
	Name     string      `orm:"size(255)"`
	Email    string      `orm:"size(255)"`
	StaffLog []*StaffLog `orm:"reverse(many)"`
}

func (m *User) GetAll() (*[]*User, error) {
	var result []*User

	o := orm.NewOrm()
	qs := o.QueryTable(m)

	_, err := qs.All(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
