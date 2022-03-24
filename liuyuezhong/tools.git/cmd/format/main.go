package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
)

type LogEntry struct {
	IP      string `json:"real_ip"`
	URL     string `json:"req_uri"`
	ReqBody string `json:"req_body"`
	RspBody string `json:"resp_body"`
	TraceId string `json:"trace_id"`

	reqBody interface{}
	rspBody struct {
		DmError int         `json:"dm_error"`
		ErrMsg  string      `json:"error_msg"`
		Data    interface{} `json:"data"`
	}
}

type KeyMsg struct {
	TraceId string      `json:"trace_id"`
	Url     string      `json:"url"`
	Uid     string      `json:"uid"`
	IP      string      `json:"real_ip"`
	Body    interface{} `json:"body"`
	DmError int         `json:"dm_error"`
	ErrMsg  string      `json:"error_msg"`
	Data    interface{} `json:"data"`
}

func TakeLog(s string) {
	var e LogEntry
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		log.Fatal(err)
		return
	}

	if err := json.Unmarshal([]byte(e.ReqBody), &e.reqBody); err != nil {
		log.Fatal(err)
		return
	}

	if err := json.Unmarshal([]byte(e.RspBody), &e.rspBody); err != nil {
		log.Fatal(err)
		return
	}

	uri, err := url.Parse(e.URL)
	if err != nil {
		log.Fatal(err)
		return
	}

	query := uri.Query()

	msg := KeyMsg{
		TraceId: e.TraceId,
		Url:     uri.Path,
		Uid:     query.Get("uid"),
		IP:      e.IP,
		Body:    e.reqBody,
		DmError: e.rspBody.DmError,
		ErrMsg:  e.rspBody.ErrMsg,
		Data:    e.rspBody.Data,
	}

	data, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(string(data))
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			break
		}
		TakeLog(text)
	}
}
