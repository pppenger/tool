package main

import (
	"github.com/pquerna/otp/totp"
	"log"
	"time"
)

func Totp(s string) string {
	value, err := totp.GenerateCode(s, time.Now())
	if err != nil {
		log.Fatalf("totp.GenerateCode err: %v", err)
	}

	return value
}

func GetSecret(t, secret string) string {
	if t == "totp" {
		return Totp(secret)
	}

	return secret
}
