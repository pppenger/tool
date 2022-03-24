package main

import "time"

type UidRegisteredCv struct {
	ID         int64     `gorm:"column:id;primary_key"`
	UID        int64     `gorm:"column:uid"`
	CreateTime time.Time `gorm:"column:create_time"`
	CvPrefix   string    `gorm:"column:cv_prefix"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (UidRegisteredCv) TableName() string {
	return "uid_registered_cv"
}

type UserPhone struct {
	ID         int64     `gorm:"column:id;primary_key"`
	UID        int64     `gorm:"column:uid"`
	Phone      string    `gorm:"column:phone"`
	Stat       int64     `gorm:"column:stat"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Cv         string    `gorm:"column:cv"`
	CvPrefix   string    `gorm:"column:cv_pre"`
	AppID      int64     `gorm:"column:appid"`
}

func (UserPhone) TableName() string {
	return "user_phone"
}
