package main

import (
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
	if err := source.FromTxt(file, func(records [][]string) {
		for _, record := range records {
			line := record[0]
			if debug {
				fmt.Printf("parse line: %s\n", line)
			}

			msg, err := ParseToFinanceMsg(line)
			if err != nil {
				log.Fatalf("source.FromTxt err: %v", err)
			}

			if msg == nil {
				continue
			}

			db, tb := GetDbTbName(msg.SendMsgid)
			log.Printf("insert into [%s].[%s]: %d\n", db, tb, msg.SendMsgid)

			if insert {
				c := dao.GetDB(db).Table(tb)
				var results []FinanceMsg
				if err := c.Where("send_msgid = ? and billid = ?", msg.SendMsgid, msg.BillId).Find(&results).Error; err != nil {
					log.Fatalf("sql find err: %v", err)
				}

				if len(results) != 0 {
					if len(results) == 1 {
						log.Printf("skip sendMsgId: %d\n", msg.SendMsgid)
					} else if len(results) > 1 {
						log.Printf("fuck sendMsgId: %d, has count %d\n", msg.SendMsgid, len(results))
						deleteIds := make([]uint64, 0, len(results)-1)
						var max uint64
						for _, r := range results {
							if r.Id > max {
								max = r.Id
							}
						}

						for _, r := range results {
							if r.Id != max {
								deleteIds = append(deleteIds, r.Id)
							}
						}

						if len(deleteIds)+1 != len(results) {
							log.Fatalf("fuck msgId: %d", msg.SendMsgid)
						} else {
							if err := c.Where("id in (?) and send_msgid = ?", deleteIds, msg.SendMsgid).Delete(FinanceMsg{}).Error; err != nil {
								log.Fatalf("del msgId: %d", msg.SendMsgid)
							}
						}
						log.Printf("clean ok for msgId: %d\n", msg.SendMsgid)
					}

					continue
				}

				if msg.Id != 0 {
					msg.Id = 0
					log.Printf("msg id != 0, msg: %+v", msg)
				}
				if err := dao.GetDB(db).Table(tb).Create(msg).Error; err != nil {
					log.Fatalf("sql create err: %v, msg: %+v", err, msg)
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
