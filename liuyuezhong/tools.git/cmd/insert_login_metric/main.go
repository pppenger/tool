package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/popstk/inke/internal/config"
	"github.com/popstk/inke/internal/dao"
	"github.com/popstk/inke/internal/source"
	"log"
)

var (
	configFile string
	insert     bool
	debug      bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	flag.StringVar(&configFile, "c", "connect.toml", "connect file path")
	flag.BoolVar(&insert, "i", false, "do insert")
	flag.BoolVar(&debug, "d", false, "debug mode")
}

func main() {
	flag.Parse()

	fmt.Printf("load config %s\n", configFile)
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("config.LoadConfig %s err: %v", configFile, err)
	}

	d, err := dao.NewDao(conf)
	if err != nil {
		log.Fatalf("dao.NewDao err: %v", err)
	}

	for _, file := range flag.Args() {
		log.Println("handle file ", file)
		if err := handleFile(d, file); err != nil {
			log.Fatalf("handleFile err: %v", err)
		}
	}
}

func handleFile(dao *dao.Dao, file string) error {
	db := dao.GetDB("alarm")

	if err := source.FromTxt(file, func(records [][]string) {
		for _, record := range records {
			line := record[0]
			if debug {
				fmt.Printf("parse line: %s\n", line)
			}

			msg, err := ParseToStruct(line)
			if err != nil {
				log.Fatalf("source.FromTxt err: %v", err)
			}

			if msg == nil {
				continue
			}

			log.Printf("insert msg: %+v\n", msg.Atom)
			if insert {
				wechatStr, _ := json.Marshal(msg.WechatLogin)
				atomStr, _ := json.Marshal(msg.Atom)

				record := LoginRequest{
					LoginPath:   msg.LoginPath,
					ErrCode:     msg.ErrCode,
					EventTime:   msg.EventTime,
					Smid:        msg.Atom.Smid,
					Cv:          msg.Atom.Cv,
					Phone:       msg.Phone,
					Uid:         msg.Uid,
					WechatLogin: string(wechatStr),
					Atom:        string(atomStr),
				}

				if err := db.Create(&record).Error; err != nil {
					log.Fatalf("sql create err: %v", err)
				}
			}
		}
	}, func(s string) []string {
		return []string{s}
	}); err != nil {
		return fmt.Errorf("source.FromTxt err: %v", err)
	}

	return nil
}
