package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"regexp"
	"strings"
)

type QASubject struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
	Right    int      `json:"right"`
}

func importQA(c *redis.Client, dataFile string) {
	qR := regexp.MustCompile(`^(\d*)\.*`)
	aR := regexp.MustCompile(`^\s*[ABCD]\.*\s*`)

	ScanExcelLines(dataFile, 2, func(l int, records []string) bool {
		if len(records) != 6 {
			fmt.Printf("invalid data: %v\n", records)
			return false
		}

		q := qR.ReplaceAllString(strings.TrimSpace(records[0]), "")

		var answers []string
		for i := 1; i < 5; i++ {
			answers = append(answers,
				aR.ReplaceAllString(strings.TrimSpace(records[i]), ""))
		}

		right := strings.TrimSpace(records[5])
		if len(right) == 0 {
			fmt.Printf("invalid right answer %+v\n", records)
			return true
		}
		if right[0] < 'A' || right[0] > 'D' {
			fmt.Printf("invalid right answer %+v\n", records)
			return true
		}

		i := right[0] - 'A'
		s := QASubject{
			Question: q,
			Answers:  answers,
			Right:    int(i),
		}
		data, err := json.Marshal(s)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("[%d]data: %s\n", l, string(data))
		c.RPush("red_package:qa", data)

		return true
	})
}
