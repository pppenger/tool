package main

import (
	"fmt"
	"github.com/popstk/inke/internal/format"
	"regexp"
	"time"
)

type LoginRequest struct {
	ID          int       `json:"id" gorm:"column:id"`
	LoginPath   LoginPath `json:"login_path" gorm:"column:login_path"` // 请求类型
	ErrCode     int       `json:"err_code" gorm:"column:err_code"`     // 错误码
	EventTime   time.Time `json:"event_time" gorm:"column:event_time"` // 创建时间
	Smid        string    `json:"smid" gorm:"column:smid"`
	Cv          string    `json:"cv" gorm:"column:cv"`
	Phone       string    `json:"phone" gorm:"column:phone"`
	Uid         uint64    `json:"uid" gorm:"column:uid"`
	WechatLogin string    `json:"wechat_login" gorm:"column:wechat_login"`
	Atom        string    `json:"atom" gorm:"column:atom"`
}

func (m *LoginRequest) TableName() string {
	return "login_request"
}

type LoginPath int8

const (
	LoginPathNone LoginPath = iota
	LoginPathPhoneCode
	LoginPathPhone
	LoginPathThirdPart
	LoginPathShanyan
	LoginPathEnd
)

type WechatLogin struct {
	OpenId  string `json:"openid"`
	UnionId string `json:"unionid"`
}

type LoginMetric struct {
	LoginPath   LoginPath    `json:"login_path"`
	ErrCode     int          `json:"err_code"`
	EventTime   time.Time    `json:"event_time"`
	Atom        LoginAtom    `json:"atom"`
	NewAdd      bool         `json:"new_add"`
	Phone       string       `json:"phone"`
	Uid         uint64       `json:"uid"`
	WechatLogin *WechatLogin `json:"wechat_login"`
}

type LoginAtom struct {
	FormWeb    int    `schema:"from_web" json:"from_web"`
	ClientIp   string `schema:"xrealip" json:"client_ip"`
	Uid        uint64 `schema:"uid" json:"uid"`
	Cc         string `schema:"cc" json:"cc"`
	Cv         string `schema:"cv" json:"cv"`
	Ua         string `schema:"ua" json:"ua"`
	Conn       string `schema:"conn" json:"conn"`
	Devi       string `schema:"devi" json:"devi"`
	Idfv       string `schema:"idfv" json:"idfv"`
	Idfa       string `schema:"idfa" json:"idfa"`
	Proto      string `schema:"proto" json:"proto"`
	OsVersion  string `schema:"osversion" json:"osversion"`
	LogId      string `schema:"logid" json:"logid"`
	Smid       string `schema:"smid" json:"smid"`
	WebProject string `schema:"webproject" json:"webproject"`
	Lc         string `schema:"lc" json:"lc"`
	Imei       string `schema:"imei" json:"imei"`
	Imsi       string `schema:"imsi" json:"imsi"`
	Token      string `schema:"token" json:"token"`
	Oaid       string `schema:"oaid" json:"oaid"`
	DevName    string `schema:"dev_name" json:"dev_name"` //用户设备类型
	Aid        string `json:"aid" schema:"aid"`
	Meid       string `json:"meid" schema:"meid"` //即imei,因为旧业务原因导致传参meid就是imei
}

var (
	reg = regexp.MustCompile(`handleLoginMetricMessage msg: (&{.+?})`)
)

func ParseToStruct(line string) (*LoginMetric, error) {
	results := reg.FindStringSubmatch(line)
	if len(results) != 2 {
		return nil, nil
	}

	var msg LoginMetric
	if err := format.FmtUnmarshalStruct(results[1], &msg); err != nil {
		return nil, fmt.Errorf("FmtUnmarshalStruct err: %v", err)
	}

	return &msg, nil
}
