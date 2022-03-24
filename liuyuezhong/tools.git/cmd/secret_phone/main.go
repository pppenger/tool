package main

import (
	"flag"
	"fmt"
)

var (
	checkAes bool
	checkRsa bool
)

func init() {
	flag.BoolVar(&checkAes, "aes", false, "aes")
	flag.BoolVar(&checkRsa, "aes", false, "rsa")
}

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		if checkAes {
			s, err := NewPhoneSecret(arg)
			if err != nil {
				fmt.Printf("%s -> %v \n", arg, err)
				continue
			}

			fmt.Printf("%s -> %s \n", s.Phone, s.SecretPhone)
		}

		if checkRsa {
			value, err := rsaDecrypt(arg)
			if err != nil {
				fmt.Printf("%s -> %v \n", arg, err)
			} else {
				fmt.Printf("%s -> %v \n", arg, value)
			}
		}
	}
}
