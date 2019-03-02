package Model

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type StaffLog struct {
	Id           uint64
	ResultType   string    `orm:"type(char);size(60)"`
	ResultString string    `orm:"type(text)"` // 结果
	Time         time.Time // 时间
	User         *User     `orm:"rel(fk);default(0);index;on_delete(do_nothing)"` // 用户表
}

func (m *StaffLog) CreateLog() {
	o := orm.NewOrm()
	m.Time = time.Now()
	_, _ = o.Insert(m)
}
