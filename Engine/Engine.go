package Engine

import (
	Crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	. "github.com/MalsonQu/AutoStaffCardForAiRZQ/Config"
	"github.com/MalsonQu/AutoStaffCardForAiRZQ/Email"
	"github.com/MalsonQu/AutoStaffCardForAiRZQ/Model"
	"github.com/franela/goreq"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var pubKey interface{}

// 用户信息数据结构
type user struct {
	UserName    string
	WaitTime    int64
	RequestBody *requestBody
	goMail      *Email.Email
}

// 请求体数据结构
type requestBody struct {
	OpenId  string `json:"openid"`  // 微信openId
	UserId  uint64 `json:"userid"`  // 用户ID
	Date    string `json:"date"`    // 签到日期
	Lng     string `json:"jingdu"`  // 精度
	Lat     string `json:"weidu"`   // 纬度
	Address string `json:"address"` // 地址文字
}

// 地址数据结构
type locationRange struct {
	maxLng float64 // 最大经度
	maxLat float64 // 最大纬度
	minLng float64 // 最小经度
	minLat float64 // 最小纬度
}

type staffRequestDataCheckData struct {
	TypeExplain string `json:"type_explain"`
	StatusInfo  string `json:"status_info"`
}

type staffRequestData struct {
	Time      string                    `json:"time"`
	CheckData staffRequestDataCheckData `json:"check_data"`
}

type staffRequest struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    staffRequestData `json:"data"`
}

// 等待
var wg = sync.WaitGroup{}

// 经纬度范围
var _locationRange locationRange

// 构建经纬度范围
func buildLocationRange() {
	earthR := 6371.0
	dLng := (2 * math.Asin(math.Sin(Config.Global.PositionRange/(2*earthR))/math.Cos(Config.Global.BaseLat*math.Pi/180))) * 180 / math.Pi
	dLat := (Config.Global.PositionRange / earthR) * 180 / math.Pi

	_locationRange = locationRange{
		maxLng: Config.Global.BaseLng + dLng,
		maxLat: Config.Global.BaseLat + dLat,
		minLng: Config.Global.BaseLng - dLng,
		minLat: Config.Global.BaseLat - dLat,
	}
}

// 随机纬度
func randLat(minLat, maxLat float64) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return strconv.FormatFloat((r.Float64()*(maxLat-minLat))+minLat, 'f', 11, 64)
}

// 随机经度
func randLng(minLng, maxLng float64) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return strconv.FormatFloat((r.Float64()*(maxLng-minLng))+minLng, 'f', 11, 64)
}

// 随机等待时间
func randWaitTime() int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Int63n(Config.Global.WaitTime.Max-Config.Global.WaitTime.Min) + Config.Global.WaitTime.Min
}

// 构建地址文字信息
func buildAddress(user *user) {
	user.RequestBody.Lng = randLng(_locationRange.minLng, _locationRange.maxLng)
	user.RequestBody.Lat = randLat(_locationRange.minLat, _locationRange.maxLat)
	wg.Done()
}

// 生成用户信息
func getUserList() (*[]*user, error) {
	var list []*user

	userList, err := (&Model.User{}).GetAll()

	if err != nil {
		fmt.Println(err.Error(), 236)
		return nil, err
	}

	// 循环用户
	for _, value := range *userList {
		wg.Add(1)
		go func(v *Model.User) {
			// 构建用户的信息
			userInfo := user{
				UserName: v.Name,
				WaitTime: randWaitTime(),
				RequestBody: &requestBody{
					UserId:  v.Id,
					OpenId:  v.WeChatId,
					Address: Config.Global.DefaultAddress,
				},
				goMail: Email.New().SetTo(v.Email),
			}
			list = append(list, &userInfo)
			wg.Done()
		}(value)
	}

	return &list, nil
}

// RSA加密
func RsaEncrypt(originData string) (string, error) {
	encryptedData, err := rsa.EncryptPKCS1v15(Crand.Reader, pubKey.(*rsa.PublicKey), []byte(originData))
	return base64.StdEncoding.EncodeToString(encryptedData), err
}

func paramEncrypt(openId, userId, lng, lat string) (string, error) {
	param, err := json.Marshal(map[string]string{
		"openid": openId,
		"userid": userId,
		"jingdu": lng,
		"weidu":  lat,
	})

	if err != nil {
		return "", err
	}

	paramData := string(param)

	return RsaEncrypt(paramData)
}

// 打卡
func staffCard(user *user) {
	// 获取当前时间
	_funcStartTime := time.Now().Unix()
	// 设置时间种子
	_timeSeed := time.Unix(_funcStartTime+user.WaitTime, 0)
	// 获取时间对象
	user.RequestBody.Date = _timeSeed.Format("2006年01月02日 15:04:05 ")
	//
	mailData := _timeSeed.Format("15:04:05")

	wg.Add(1)
	go func() {
		// 发送确认邮件
		user.goMail.StaffConfirm(mailData).Send()
		wg.Done()
	}()

	// 设置睡眠时间
	_sleepTimeDifference := time.Now().Unix() - _funcStartTime

	// 先睡下
	time.Sleep(time.Duration(user.WaitTime-_sleepTimeDifference) * time.Second)

	body := url.Values{}

	// 一会需要加密
	// 构建需要加密数据
	param, err := paramEncrypt(
		user.RequestBody.OpenId,
		strconv.FormatUint(user.RequestBody.UserId, 10),
		user.RequestBody.Lng,
		user.RequestBody.Lat,
	)

	// json 解析错误
	if err != nil {
		fmt.Println(err.Error(), 306)
		user.goMail.StaffFailed("打卡失败-" + err.Error()).Send()
		(&Model.StaffLog{
			ResultType:   "FAILED",
			ResultString: "打卡失败-" + err.Error(),
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
		return
	}

	body.Add("param", param)
	body.Add("userok", "0")
	body.Add("address", user.RequestBody.Address)

	// 开始打卡
	request := goreq.Request{
		Uri:         "https://hr.zihai.cn/index.php/interfaces/punchcard/staffcard",
		Method:      "POST",
		Timeout:     60 * time.Second,
		Body:        body.Encode(),
		ContentType: "application/x-www-form-urlencoded; charset=UTF-8",
		Host:        "hr.zihai.cn",
		Accept:      "*/*",
		UserAgent:   "Mozilla/5.0 (Linux; Android 8.0; G8142 Build/47.1.A.16.20; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/57.0.2987.132 MQQBrowser/6.2 TBS/044306 Mobile Safari/537.36 MicroMessenger/6.7.3.1360(0x26070337) NetType/WIFI Language/zh_CN Process/tools",
	}

	request.AddHeader("X-Requested-With", "com.tencent.mm")
	request.AddHeader("Referer", "http://hr.zihai.cn/personal/daka.php")
	request.AddHeader("Accept-Language", "zh-CN,en-US;q=0.9")

	req, err := request.Do()

	if err != nil {
		fmt.Println(err.Error(), 306)
		user.goMail.StaffFailed("打卡失败-" + err.Error()).Send()
		(&Model.StaffLog{
			ResultType:   "FAILED",
			ResultString: "打卡失败-" + err.Error() + "-" + fmt.Sprintf("%+v", *user.RequestBody),
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
		return
	}

	defer func(s *goreq.Body) {
		_ = s.Close()
	}(req.Body)

	var jsonStaffRequest staffRequest

	err = req.Body.FromJsonTo(&jsonStaffRequest)

	if err != nil {
		fmt.Println(err.Error(), 328)
		user.goMail.StaffFailed("打卡失败-" + err.Error()).Send()
		(&Model.StaffLog{
			ResultType:   "FAILED",
			ResultString: "打卡失败-" + err.Error(),
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
		return
	}

	if jsonStaffRequest.Code != 200 {
		user.goMail.StaffFailed("打卡失败-" + jsonStaffRequest.Message).Send()
		(&Model.StaffLog{
			ResultType:   "FAILED",
			ResultString: "打卡失败-" + jsonStaffRequest.Message + "-" + fmt.Sprintf("%+v", *user.RequestBody),
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
		return
	}

	// 发送打卡成功的邮件
	wg.Add(1)
	go func() {
		//取消发送打卡成功的邮件
		_staffTime := time.Now().Format("15:04:05")
		user.goMail.StaffSuccess(_staffTime, jsonStaffRequest.Data.CheckData.StatusInfo, jsonStaffRequest.Data.CheckData.TypeExplain).Send()
		// 记录数据库
		(&Model.StaffLog{
			ResultType:   "SUCCESS",
			ResultString: jsonStaffRequest.Data.CheckData.TypeExplain,
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
	}()

	wg.Done()
}

func init() {
	buildLocationRange()
}

func Run() {
	// 构建用户列表
	userList, err := getUserList()

	// 生成公钥
	key, _ := base64.StdEncoding.DecodeString(Config.Global.PublicKey)
	pubKey, _ = x509.ParsePKIXPublicKey(key)

	wg.Wait()

	if err != nil {
		fmt.Println(err.Error(), 384)
		Email.New().StaffFailed("构建用户列表出错").SetTo(Config.Email.MasterEmail).Send()
		return
	}

	// 开始执行
	for _, value := range *userList {
		wg.Add(1)
		go buildAddress(value)
	}
	wg.Wait()

	for _, value := range *userList {
		wg.Add(1)
		go staffCard(value)
	}
	wg.Wait()
}
