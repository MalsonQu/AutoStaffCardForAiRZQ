package Engine

import (
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

type user struct {
	UserName    string
	WaitTime    int64
	RequestBody *requestBody
	goMail      *Email.Email
}

type requestBody struct {
	OpenId  string `json:"openid"`  // 微信openId
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
	//user.RequestBody.Address = Config.Global.DefaultAddress
	//queryString := requestQueryString{
	//	Config.DbMap.Ak,
	//	user.RequestBody.Lat + "," + user.RequestBody.Lng,
	//	"json",
	//	"1",
	//	"100",
	//}
	//
	//res, err := goreq.Request{
	//	Uri:         "http://api.map.baidu.com/geocoder/v2/",
	//	Method:      "GET",
	//	QueryString: queryString,
	//	Accept:      "application/json",
	//	ContentType: "application/json",
	//	Timeout:     10 * time.Second,
	//}.Do()
	//
	//if err != nil {
	//	user.goMail.StaffFailed("经纬度换取地址文字失败").Send()
	//	(&Model.StaffLog{
	//		ResultType:   "FAILED",
	//		ResultString: "经纬度换取地址文字失败",
	//		User: &Model.User{
	//			Id: user.RequestBody.UserId,
	//		},
	//	}).CreateLog()
	//	wg.Done()
	//	return
	//}
	//
	//defer func(s *goreq.Body) {
	//	_ = s.Close()
	//}(res.Body)
	//
	//var mapResultJson mapResult
	//
	//err = res.Body.FromJsonTo(&mapResultJson)
	//
	//var _addressStatus int
	//
	//if err != nil {
	//	_addressStatus = 103
	//}else{
	//	_addressStatus = mapResultJson.Status
	//}
	//
	//var _status string
	//if _addressStatus != 0 {
	//	switch _addressStatus {
	//	case 1:
	//		_status = "百度服务器出错-" + "服务器内部错误"
	//		break
	//	case 2:
	//		_status = "百度服务器出错-" + "请求参数非法"
	//		break
	//	case 3:
	//		_status = "百度服务器出错-" + "权限校验失败"
	//		break
	//	case 4:
	//		_status = "百度服务器出错-" + "配额校验失败"
	//		break
	//	case 5:
	//		_status = "百度服务器出错-" + "ak不存在或者非法"
	//		break
	//	case 101:
	//		_status = "百度服务器出错-" + "服务禁用"
	//		break
	//	case 102:
	//		_status = "百度服务器出错-" + "不通过白名单或者安全码不对"
	//		break
	//	default:
	//		if _addressStatus >= 200 && _addressStatus < 300 {
	//			_status = "百度服务器出错-" + "无权限"
	//		} else if _addressStatus >= 300 {
	//			_status = "百度服务器出错-" + "配额错误"
	//		} else {
	//			_status = "百度服务器出错-" + "未知错误"
	//		}
	//	}
	//	// 发送邮件
	//	wg.Add(1)
	//	go func() {
	//		user.goMail.StaffFailed(_status).Send()
	//		(&Model.StaffLog{
	//			ResultType:   "FAILED",
	//			ResultString: _status,
	//			User: &Model.User{
	//				Id: user.RequestBody.UserId,
	//			},
	//		}).CreateLog()
	//		wg.Done()
	//	}()
	//	wg.Done()
	//	return
	//}
	//
	//if len(mapResultJson.Result.Pois) <= 0 {
	//	user.RequestBody.Address = Config.Global.DefaultAddress
	//} else {
	//	user.RequestBody.Address = mapResultJson.Result.FormattedAddress + mapResultJson.Result.Pois[0].Name
	//}

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

	for _, value := range *userList {
		wg.Add(1)
		go func(v *Model.User) {
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

// 打卡
func staffCard(user *user) {
	_funcStartTime := time.Now().Unix()
	_timeSeed := time.Unix(_funcStartTime+user.WaitTime, 0)
	user.RequestBody.Date = _timeSeed.Format("2006年01月02日 15:04:05")
	mailData := _timeSeed.Format("15:04:05")

	wg.Add(1)
	go func() {
		user.goMail.StaffConfirm(mailData).Send()
		wg.Done()
	}()

	_sleepTimeDifference := time.Now().Unix() - _funcStartTime

	// 先睡下
	time.Sleep(time.Duration(user.WaitTime-_sleepTimeDifference) * time.Second)

	body := url.Values{}

	body.Add("openid", user.RequestBody.OpenId)
	body.Add("userid", strconv.FormatUint(user.RequestBody.UserId, 10))
	body.Add("date", user.RequestBody.Date)
	body.Add("jingdu", user.RequestBody.Lng)
	body.Add("weidu", user.RequestBody.Lat)
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
		//_staffTime := time.Now().Format("15:04:05")
		//user.goMail.StaffSuccess(_staffTime, jsonStaffRequest.Data.CheckData.StatusInfo, jsonStaffRequest.Data.CheckData.TypeExplain).Send()
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
