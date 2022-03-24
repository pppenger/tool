package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"time"
)

type RegistStaticV3 struct {
	ID          int64     `json:"id" gorm:"column:id"`
	Uid         int64     `json:"uid" gorm:"column:uid"`                     // 用户id
	RegistIp    string    `json:"regist_ip" gorm:"column:regist_ip"`         // 注册ip
	RegistTime  time.Time `json:"regist_time" gorm:"column:regist_time"`     // 注册时间
	RegistCc    string    `json:"regist_cc" gorm:"column:regist_cc"`         // 注册cc
	RegistCvPre string    `json:"regist_cv_pre" gorm:"column:regist_cv_pre"` // 注册cv前缀
	RegistCv    string    `json:"regist_cv" gorm:"column:regist_cv"`         // 注册cv
	RegistSmid  string    `json:"regist_smid" gorm:"column:regist_smid"`     // 注册smid
	DevName     string    `json:"dev_name" gorm:"column:dev_name"`           // 注册设备
	Osversion   string    `json:"osversion" gorm:"column:osversion"`         // 注册系统版本
}

func (m *RegistStaticV3) TableName() string {
	return "regist_static_v3"
}

type RegistStaticV3Bak RegistStaticV3

func (m *RegistStaticV3Bak) TableName() string {
	return "regist_static_v3_modify_only_20211228"
}

// 注册uid atom替换
func importAtom(dataFile string) {
	db := GetDB()
	r := GetRedis()
	ScanExcelLines(dataFile, 2, func(i int, records []string) bool {
		if len(records) != 10 {
			fmt.Printf("invalid data: %v\n", records)
			return false
		}

		uid, err := strconv.ParseInt(records[0], 10, 64)
		if err != nil {
			fmt.Printf("invalid data: %v\n", records)
			return false
		}

		if uid == 0 {
			fmt.Printf("uid == 0, invalid data: %v\n", records)
			return false
		}

		if err := r.Del(fmt.Sprintf("USER_REGISTER:%d", uid)).Err(); err != nil {
			fmt.Printf("redis del cache uid %d err: %v\n", uid, err)
			return false
		}

		var r RegistStaticV3
		if err := db.Model(&r).Where("uid = ?", uid).First(&r).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return true
			}

			fmt.Printf("find uid = %d, err: %v", uid, err)
			return false
		}

		cv := strings.TrimSpace(records[5])
		cc := strings.TrimSpace(records[6])
		smid := strings.TrimSpace(records[7])
		osVersion := strings.TrimSpace(records[9])

		modify := make(map[string]interface{})
		if r.RegistCc != cc {
			modify["regist_cc"] = cc
		}
		if r.RegistCv != cv {
			modify["regist_cv"] = cv
			modify["regist_cv_pre"] = GetCvPrefix(cv)
		}
		if r.RegistSmid != smid {
			modify["regist_smid"] = smid
		}
		if r.Osversion != osVersion {
			modify["osversion"] = osVersion
		}

		if len(modify) == 0 {
			return true
		}

		fmt.Printf("[%d] modify uid = %d -> %+v \n", i, uid, modify)

		if !test {
			old := RegistStaticV3Bak(r)
			if err := db.Model(old).Create(&old).Error; err != nil {
				fmt.Printf("backup err: %v", err)
				return false
			}
			if err := db.Model(&r).Where("uid = ?", r.Uid).Updates(modify).Error; err != nil {
				fmt.Printf("updates err: %v", err)
				return false
			}
		}

		if test && i%10 == 0 {
			return false
		}

		return true
	})
}
