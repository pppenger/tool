package main

import (
	"fmt"
	"github.com/popstk/inke/internal/format"
	"regexp"
	"time"
)

type FinanceMsg struct {
	Id              uint64    `json:"id" gorm:"PRIMARY_KEY;AUTO_INCREMENT;column:id"`
	SendUid         uint64    `json:"send_uid" gorm:"column:send_uid"`
	SendMsgid       uint64    `json:"send_msgid" gorm:"column:send_msgid"`
	SendCreateTime  time.Time `json:"send_create_time" gorm:"column:send_create_time"`
	BillId          string    `json:"billid" gorm:"column:billid"`
	WithdrawBillId  string    `json:"withdraw_billid" gorm:"column:withdraw_billid"`
	ReplyBillId     string    `json:"reply_billid" gorm:"column:reply_billid"`
	Status          int       `json:"status" gorm:"column:status"` //0-待回复 1-已回复 2-已过期
	ReplyUid        uint64    `json:"reply_uid" gorm:"column:reply_uid"`
	ReplyMsgid      uint64    `json:"reply_msgid" gorm:"column:reply_msgid"`
	ReplyCreateTime time.Time `json:"reply_create_time" gorm:"column:reply_create_time"`
	Coins           int       `json:"coins" gorm:"column:coins"`
	Points          int       `json:"points" gorm:"column:points"`
	IsSysSend       int       `json:"is_sys_send" gorm:"column:is_sys_send"`
	MsgType         int32     `json:"msg_type" gorm:"column:msg_type"`
	ChatUpGiftId    int       `json:"chatup_giftid" gorm:"column:chatup_giftid"`
}

var (
	reg = regexp.MustCompile(`mysql error: msg (&{.+?})`)
)

func ParseToFinanceMsg(line string) (*FinanceMsg, error) {
	results := reg.FindStringSubmatch(line)
	if len(results) != 2 {
		return nil, nil
	}

	var msg FinanceMsg
	if err := format.FmtUnmarshalStruct(results[1], &msg); err != nil {
		return nil, fmt.Errorf("FmtUnmarshalStruct err: %v", err)
	}

	return &msg, nil
}

const (
	DbDenominator = 1000
	DbMember      = 100
	TbMember      = 100

	DbNameFmt = "message_%d"
	TbNameFmt = "finance_msg_%d"
)

func GetDbTbName(msgid uint64) (string, string) {
	dbSlot := msgid % DbDenominator / DbMember
	tbSlot := msgid % TbMember
	dbName := fmt.Sprintf(DbNameFmt, dbSlot)
	tbName := fmt.Sprintf(TbNameFmt, tbSlot)
	return dbName, tbName
}
