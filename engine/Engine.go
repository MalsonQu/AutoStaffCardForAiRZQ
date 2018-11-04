package engine

import (
	"autoStaffCardForAiRZQ/Model"
	"autoStaffCardForAiRZQ/email"
	"fmt"
	"github.com/franela/goreq"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type user struct {
	UserName    string
	WaitTime    int64
	RequestBody *requestBody
	goMail      *email.Email
}

type requestBody struct {
	UserId  uint64 `json:"userid"`  // 用户ID
	Date    string `json:"date"`    // 签到日期
	Lng     string `json:"jingdu"`  // 精度
	Lat     string `json:"weidu"`   // 纬度
	Address string `json:"address"` // 地址文字
}

type locationRange struct {
	maxLng float64 // 最大经度
	maxLat float64 // 最大纬度
	minLng float64 // 最小经度
	minLat float64 // 最小纬度
}

type waitTimeRange struct {
	min int64
	max int64
}

type requestQueryString struct {
	Ak       string `url:"ak"`
	Location string `url:"location"`
	Output   string `url:"output"`
	Pois     string `url:"pois"`
	Radius   string `url:"radius"`
}

type mapResultPois struct {
	Name string `json:"name"`
}

type mapResultContent struct {
	FormattedAddress string          `json:"formatted_address"`
	Pois             []mapResultPois `json:"pois"`
}

type mapResult struct {
	Status int              `json:"status"`
	Result mapResultContent `json:"result"`
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

var jsonStaffRequest staffRequest

// 基础经度
var _baseLng = 126.649806

// 基础纬度
var _baseLat = 45.782213

// 经纬度范围
var _locationRange locationRange

// 时间范围
var _waitTimeRange waitTimeRange

// 等待
var wg = sync.WaitGroup{}

// 构建经纬度范围
func buildLocationRange() {
	earthR := 6371.0
	dis := 0.01
	dLng := (2 * math.Asin(math.Sin(dis/(2*earthR))/math.Cos(_baseLat*math.Pi/180))) * 180 / math.Pi
	dLat := (dis / earthR) * 180 / math.Pi

	_locationRange = locationRange{
		maxLng: _baseLng + dLng,
		maxLat: _baseLat + dLat,
		minLng: _baseLng - dLng,
		minLat: _baseLat - dLat,
	}
}

// 构建等待时间
func buildWaitTimeRange() {
	_waitTimeRange = waitTimeRange{
		min: 1,
		max: 120,
	}
}

// 构建维度和纬度
func buildLocation(user *user) {
	user.RequestBody.Lat = randLat()
	user.RequestBody.Lng = randLng()
	// 构建地址
	buildAddress(user)
}

// 随机纬度
func randLat() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return strconv.FormatFloat((r.Float64()*(_locationRange.maxLat-_locationRange.minLat))+_locationRange.minLat, 'f', 11, 64)
}

// 随机经度
func randLng() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return strconv.FormatFloat((r.Float64()*(_locationRange.maxLng-_locationRange.minLng))+_locationRange.minLng, 'f', 11, 64)
}

// 随机等待时间
func randWaitTime() int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Int63n(_waitTimeRange.max-_waitTimeRange.min) + _waitTimeRange.min
}

// 构建地址文字信息
func buildAddress(user *user) {
	queryString := requestQueryString{
		"",
		user.RequestBody.Lat + "," + user.RequestBody.Lng,
		"json",
		"1",
		"100",
	}

	res, err := goreq.Request{
		Uri:         "http://api.map.baidu.com/geocoder/v2/",
		Method:      "GET",
		QueryString: queryString,
		Accept:      "application/json",
		ContentType: "application/json",
		Timeout:     10 * time.Second,
	}.Do()

	if err != nil {
		user.goMail.StaffFailed("经纬度换取地址文字失败").Send()
		(&Model.StaffLog{
			ResultType:   "FAILED",
			ResultString: "经纬度换取地址文字失败",
			User: &Model.User{
				Id: user.RequestBody.UserId,
			},
		}).CreateLog()
		wg.Done()
		return
	}

	defer res.Body.Close()

	var mapResultJson mapResult

	res.Body.FromJsonTo(&mapResultJson)

	_addressStatus := mapResultJson.Status

	var _status string
	if _addressStatus != 0 {
		switch _addressStatus {
		case 1:
			_status = "百度服务器出错-" + "服务器内部错误"
			break
		case 2:
			_status = "百度服务器出错-" + "请求参数非法"
			break
		case 3:
			_status = "百度服务器出错-" + "权限校验失败"
			break
		case 4:
			_status = "百度服务器出错-" + "配额校验失败"
			break
		case 5:
			_status = "百度服务器出错-" + "ak不存在或者非法"
			break
		case 101:
			_status = "百度服务器出错-" + "服务禁用"
			break
		case 102:
			_status = "百度服务器出错-" + "不通过白名单或者安全码不对"
			break
		default:
			if _addressStatus >= 200 && _addressStatus < 300 {
				_status = "百度服务器出错-" + "无权限"
			} else if _addressStatus >= 300 {
				_status = "百度服务器出错-" + "配额错误"
			} else {
				_status = "百度服务器出错-" + "未知错误"
			}
		}
		// 发送邮件
		go func() {
			wg.Add(1)
			user.goMail.StaffFailed(_status).Send()
			(&Model.StaffLog{
				ResultType:   "FAILED",
				ResultString: _status,
				User: &Model.User{
					Id: user.RequestBody.UserId,
				},
			}).CreateLog()
			wg.Done()
		}()
		wg.Done()
		return
	}

	user.RequestBody.Address = mapResultJson.Result.FormattedAddress + mapResultJson.Result.Pois[0].Name
	wg.Done()
}

// 生成用户信息
func buildUserList() (*[]*user, error) {
	var list []*user

	userList, err := (&Model.User{}).GetAll()

	if err != nil {
		return nil, err
	}

	for _, value := range *userList {
		userInfo := user{
			UserName: value.Name,
			WaitTime: randWaitTime(),
			RequestBody: &requestBody{
				UserId: value.Id,
			},
			goMail: email.New().SetTo(value.Email),
		}
		list = append(list, &userInfo)
	}

	return &list, nil
}

// 打卡
func staffCard(user *user) {
	_funcStartTime := time.Now().Unix()
	_timeSeed := time.Unix(_funcStartTime+user.WaitTime, 0)
	user.RequestBody.Date = _timeSeed.Format("15:04")
	mailData := _timeSeed.Format("15:04:05")

	go func() {
		wg.Add(1)
		user.goMail.StaffConfirm(mailData).Send()
		wg.Done()
	}()

	_sleepTimeDifference := time.Now().Unix() - _funcStartTime

	// 先睡下
	time.Sleep(time.Duration(user.WaitTime-_sleepTimeDifference) * time.Second)

	body := url.Values{}

	body.Add("userid", strconv.FormatUint(user.RequestBody.UserId, 10))
	body.Add("date", user.RequestBody.Date)
	body.Add("jingdu", user.RequestBody.Lng)
	body.Add("weidu", user.RequestBody.Lat)
	body.Add("address", user.RequestBody.Address)

	// 开始打卡
	request := goreq.Request{
		Uri:         "http://hr.zihai.cn/index.php/interfaces/punchcard/staffcard",
		Method:      "POST",
		Timeout:     20 * time.Second,
		Body:        body.Encode(),
		ContentType: "application/x-www-form-urlencoded; charset=UTF-8",
		Host:        "hr.zihai.cn",
		Accept:      "*/*",
		UserAgent:   "Mozilla/5.0 (Linux; Android 8.0; G8142 Build/47.1.A.16.20; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/57.0.2987.132 MQQBrowser/6.2 TBS/044306 Mobile Safari/537.36 MicroMessenger/6.7.3.1360(0x26070337) NetType/WIFI Language/zh_CN Process/tools",
	}

	request.AddHeader("X-Requested-With", "XMLHttpRequest")
	request.AddHeader("Referer", "http://hr.zihai.cn/personal/daka.php")
	request.AddHeader("Accept-Language", "zh-CN,zh-CN;q=0.8,en-US;q=0.6")

	req, err := request.Do()

	if err != nil {
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

	defer req.Body.Close()

	req.Body.FromJsonTo(&jsonStaffRequest)

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

	_staffTime := time.Now().Format("15:04:05")
	// 发送打卡成功的邮件
	go func() {
		wg.Add(1)
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
	// 生成经纬度范围
	buildLocationRange()
	// 生成时间范围
	buildWaitTimeRange()
}

func Run() {
	// 构建用户列表
	userList, err := buildUserList()

	if err != nil {
		email.New().StaffFailed("构建用户列表出错").SetTo("").Send()
		return
	}

	// 开始执行
	for _, value := range *userList {
		wg.Add(1)
		go buildLocation(value)
	}
	wg.Wait()

	for _, value := range *userList {
		wg.Add(1)
		go staffCard(value)
	}
	wg.Wait()

}
