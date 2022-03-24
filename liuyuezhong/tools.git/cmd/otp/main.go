package main

import (
	"flag"
	"fmt"
	"github.com/pquerna/otp/totp"
	"time"
)

var (
	interval bool
)

func init() {
	flag.BoolVar(&interval, "i", false, "interval output")
}

func MustString(v string, err error) string {
	if err != nil {
		panic(err)
	}

	return v
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("need otp secret")
		return
	}

	output := func() {
		secret := flag.Arg(0)
		fmt.Print(MustString(totp.GenerateCode(secret, time.Now())))
	}

	output()

	tick := time.NewTicker(1 * time.Second)
	if !interval {
		return
	}

	for {
		fmt.Println("")
		select {
		case <-tick.C:
			output()
		}
	}
}
