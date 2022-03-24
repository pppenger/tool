package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Column int8

const (
	ColumnUid        Column = 0
	ColumnCreateTime        = 1
	ColumnPhone             = 2
	ColumnCv                = 3
	ColumnCvPre             = 4
	ColumnAppID             = 5
	ColumnUpdateTime        = 6
	ColumnStat              = 7
)

func InsertUserPhone(db *gorm.DB, values []UserPhone) {
	const fieldNums = 8

	quesMarkString := "("
	for i := 0; i < fieldNums; i++ {
		quesMarkString += "?, "
	}
	quesMarkString = quesMarkString[:len(quesMarkString)-2] + ")"

	valueStrings := make([]string, 0, len(values))
	valueArgs := make([]interface{}, 0, len(values)*fieldNums)

	for _, value := range values {
		valueStrings = append(valueStrings, quesMarkString)
		valueArgs = append(valueArgs, value.UID)
		valueArgs = append(valueArgs, value.Phone)
		valueArgs = append(valueArgs, value.Stat)
		valueArgs = append(valueArgs, value.CreateTime)
		valueArgs = append(valueArgs, value.UpdateTime)
		valueArgs = append(valueArgs, value.Cv)
		valueArgs = append(valueArgs, value.CvPrefix)
		valueArgs = append(valueArgs, value.AppID)
	}

	stmt := fmt.Sprintf("INSERT IGNORE INTO user_phone (uid, phone, stat, create_time, update_time, cv, cv_pre, appid) VALUES %s",
		strings.Join(valueStrings, ","))
	if err := db.Exec(stmt, valueArgs...).Error; err != nil {
		fmt.Println(err)
	}
}

func PhoneRoutine(db *gorm.DB, ch chan UserPhone) {
	var values []UserPhone

	count := 0
	for value := range ch {
		values = append(values, value)
		count++
		if len(values) >= batch {
			log.Printf("[batch]inset [%v, %v] \n", count-len(values)+1, count)
			InsertUserPhone(db, values)
			values = nil
		}
	}

	if len(values) > 0 {
		log.Printf("[batch]inset [%v, %v] \n", count-len(values), count-1)
		InsertUserPhone(db, values)
	}

	log.Println("[batch]wechat routine exit")
}

func WechatRoutine(db *gorm.DB, ch chan UserPhone) {
	for value := range ch {
		var results []UserPhone

		log.Printf("[wechat]try uid = %+v, phone = %+v", value.UID, value.Phone)
		if err := db.Where("uid = ? and phone = ?",
			value.UID, value.Phone).Find(&results).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Println("[wechat]skip old record")
			} else {
				log.Println("[wechat]check error for ", err)
			}

			continue
		}

		if len(results) == 1 && results[0].Cv == "" {
			err := db.Model(&UserPhone{}).Where("uid = ? and phone = ?",
				value.UID, value.Phone).Update(map[string]interface{}{
				"cv":     value.Cv,
				"cv_pre": value.CvPrefix,
				"appid":  value.AppID,
			}).Error
			if err != nil {
				log.Println("[wechat]update one line error: ", err)
			}

			continue
		}

		found := false
		for _, result := range results {
			if result.Cv == value.Cv {
				found = true
				break
			}
		}

		if found {
			log.Println("[wechat]skip inserted")
			continue
		}

		log.Println("[wechat]insert new record")
		if err := db.Create(&value).Error; err != nil {
			log.Println("[wechat]insert one line error: ", err)
		}
	}

	log.Println("wechat routine exit")
}

func importUserPhone(db *gorm.DB, dataFile string) {
	f, err := excelize.OpenFile(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("fetch rows from sheet: ", sheet)
	rows, err := f.Rows(sheet)
	if err != nil {
		log.Fatal(err)
	}

	loc := time.Now().Location()

	phoneCh := make(chan UserPhone, 1000)
	wechatCh := make(chan UserPhone, 1000)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		PhoneRoutine(db, phoneCh)
	}()

	go func() {
		defer wg.Done()
		WechatRoutine(db, wechatCh)
	}()

	log.Println("process lines...")
	for line := 1; rows.Next(); line++ {
		record := rows.Columns()
		if len(record) == 0 || line == 1 {
			continue
		}

		// note: test
		/*
			if line > 1000 {
				break
			}
		*/

		uid, err := strconv.ParseInt(record[ColumnUid], 10, 64)
		if err != nil {
			log.Fatalf("%+v skip for uid: %+v\n", record, err)
		}

		stat, err := strconv.ParseInt(record[ColumnStat], 10, 64)
		if err != nil {
			log.Fatalf("%+v skip for stat: %+v\n", record, err)
		}

		createTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", record[ColumnCreateTime], loc)
		if err != nil {
			log.Fatalf("%+v skip for create time: %+v\n", record, err)
		}

		updateTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", record[ColumnUpdateTime], loc)
		if err != nil {
			log.Fatalf("%+v skip for update time: %+v\n", record, err)
		}

		appId, err := strconv.ParseInt(record[ColumnAppID], 10, 64)
		if err != nil {
			log.Fatalf("%+v skip for appId: %+v\n", record, err)
		}

		value := UserPhone{
			UID:        uid,
			Phone:      record[ColumnPhone],
			Stat:       stat,
			CreateTime: createTime,
			UpdateTime: updateTime,
			Cv:         record[ColumnCv],
			CvPrefix:   record[ColumnCvPre],
			AppID:      appId,
		}

		if value.Stat == 0 {
			phoneCh <- value
		} else if value.Stat == 1 {
			wechatCh <- value
		} else {
			log.Printf("invalid line: %+v \n", value)
		}
	}

	close(phoneCh)
	close(wechatCh)

	log.Println("waiting exit...")
	wg.Wait()
}
