package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

func InsertRegisteredCv(db *gorm.DB, values []UidRegisteredCv) {
	const fieldNums = 4
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
		valueArgs = append(valueArgs, value.CreateTime)
		valueArgs = append(valueArgs, value.CvPrefix)
		valueArgs = append(valueArgs, value.UpdateTime)
	}

	stmt := fmt.Sprintf("INSERT IGNORE INTO uid_registered_cv (uid, create_time, cv_prefix, update_time) VALUES %s",
		strings.Join(valueStrings, ","))
	if err := db.Exec(stmt, valueArgs...).Error; err != nil {
		fmt.Println(err)
	}
}

func importRegisteredCv(db *gorm.DB, dataFile string) {
	f, err := excelize.OpenFile(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := f.Rows(sheet)
	if err != nil {
		log.Fatal(err)
	}

	loc := time.Now().Location()
	var cvs []UidRegisteredCv
	count := 0
	for ; rows.Next(); count++ {
		record := rows.Columns()
		if len(record) == 0 {
			continue
		}

		uid, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			log.Printf("%+v skip for uid: %+v\n", record, err)
			continue
		}

		createTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", record[2], loc)
		if err != nil {
			log.Printf("%+v skip for create time: %+v\n", record, err)
			continue
		}

		updateTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", record[3], loc)
		if err != nil {
			log.Printf("%+v skip for update time: %+v\n", record, err)
			continue
		}

		prefix := GetCvPrefix(record[1])
		if prefix == "" {
			log.Printf("[%d]invalid cv -> %s\n", count, record[1])
		}

		cvs = append(cvs, UidRegisteredCv{
			UID:        uid,
			CreateTime: createTime,
			CvPrefix:   prefix,
			UpdateTime: updateTime,
		})

		if len(cvs) >= batch {
			log.Printf("inset [%v, %v] \n", count-len(cvs)+1, count)
			InsertRegisteredCv(db, cvs)
			cvs = nil
		}
	}

	if len(cvs) > 0 {
		log.Printf("inset [%v, %v] \n", count-len(cvs), count-1)
		InsertRegisteredCv(db, cvs)
	}
}
