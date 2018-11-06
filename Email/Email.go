package Email

import (
	. "autoStaffCardForAiRZQ/Config"
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

type Email struct {
	goMail *gomail.Message
}

func New() *Email {
	t := Email{gomail.NewMessage()}
	t.goMail.SetHeader("From", Config.Email.Email)
	return &t
}

func (t *Email) SetTo(emailAddress string) *Email {
	t.goMail.SetHeader("To", emailAddress)
	return t
}

func (t *Email) StaffConfirm(date string) *Email {
	t.goMail.SetHeader("Subject", "打卡预通知")
	t.goMail.SetBody("text/html", `<h3>系统将在<span style="color:red"">`+date+`</span>为您自动打卡。如在后续未收到打卡结果通知，请自行打卡，以免迟到扣款！</h3><br/><span style="color:#ccc;font-size:10px">Powered By Malson (quqingyu@live.cn)</span>`)
	return t
}

func (t *Email) StaffSuccess(time, status, typeExplain string) *Email {
	t.goMail.SetHeader("Subject", "打卡成功")
	t.goMail.SetBody("text/html", `<h3>恭喜您打卡成功</h3><br/><h4>时间: `+time+`</h4><br/><h4>状态: `+status+`</h4><br/><h4>类型: `+typeExplain+`</h4><br/><span style="color:#ccc;font-size:10px">Powered By Malson (quqingyu@live.cn)</span>`)
	return t
}

func (t *Email) StaffFailed(str string) *Email {
	t.goMail.SetHeader("Subject", "打卡失败")
	t.goMail.SetBody("text/html", `<h3>很抱歉打卡失败</h3><h4>原因：`+str+`</h4><br/><a href="http://hr.zihai.cn/personal/daka.php">手动打卡</a><br/><br/><span style="color:#ccc;font-size:10px">Powered By Malson (quqingyu@live.cn)</span>`)
	return t
}

func (t *Email) Send() {
	dialer := gomail.NewDialer(
		Config.Email.Host,
		Config.Email.Port,
		Config.Email.Email,
		Config.Email.Password)

	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(t.goMail); err != nil {
		panic(err)
	}
}
