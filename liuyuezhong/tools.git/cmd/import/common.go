package main

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Signal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sigChan
		log.Printf("get a signal %s\n", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("now...")
			return
		case syscall.SIGHUP:
		default:
		}
	}
}

func GetCvPrefix(cv string) string {
	n := len(cv)
	for i := 0; i < n; i++ {
		ch := cv[i]
		if ch >= '0' && ch <= '9' {
			if i == 0 {
				return ""
			}
			return cv[:i]
		}
	}

	return cv
}

type Config struct {
	DB    string `toml:"db"`
	Redis string `toml:"redis"`
}

func ScanExcelLines(file string, start int, handle func(int, []string) bool) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("fetch rows from sheet: ", sheet)
	rows, err := f.Rows(sheet)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; rows.Next(); i++ {
		if i < start {
			continue
		}

		record := rows.Columns()
		if !handle(i, record) {
			break
		}
	}
}
