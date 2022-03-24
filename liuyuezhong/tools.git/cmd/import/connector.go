package main

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"log"
)

func GetDB() *gorm.DB {
	db, err := gorm.Open("mysql", config.DB)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func GetRedis() *redis.Client {
	opt, err := redis.ParseURL(config.Redis)
	if err != nil {
		log.Fatal(err)
	}

	return redis.NewClient(opt)
}
