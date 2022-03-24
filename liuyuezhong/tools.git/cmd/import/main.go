package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	configFile string
	sheet      string
	batch      int
	config     Config
	test       bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&configFile, "i", "connect.toml", "")
	flag.StringVar(&sheet, "sheet", "Sheet1", "sheet name")
	flag.IntVar(&batch, "batch", 5000, "max batch insert")
	flag.BoolVar(&test, "test", false, "test only")
}

func main() {
	flag.Parse()

	log.Println("start ...")
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal(err)
	}
	go Signal()

	dataFile := flag.Arg(0)
	log.Println("import file: ", dataFile)

	//importRegisteredCv(db, dataFile)
	//importUserPhone(GetDB(), dataFile)
	//importQA(GetRedis(), dataFile)
	importAtom(dataFile)

	log.Println("done!")
}
