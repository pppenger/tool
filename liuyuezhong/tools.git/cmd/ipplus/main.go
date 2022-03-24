package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ipplus360/awdb-golang/awdb-golang"
	"net"
	"net/http"
)

type Result struct {
	IP     string            `json:"ip"`
	Result map[string]string `json:"result"`
}

func ErrRsp(w http.ResponseWriter, err error) {
	w.Write([]byte(err.Error()))
}

func main() {
	ipDB, err := awdb.Open("C:\\Users\\ther\\Downloads\\cn_ip.awdb")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		value := query.Get("ip")

		ip := net.ParseIP(value)
		var record interface{}
		if err := ipDB.Lookup(ip, &record); err != nil {
			ErrRsp(w, err)
			return
		}
		resMap, ok := record.(map[string]interface{})
		if !ok {
			ErrRsp(w, errors.New("not map"))
			return
		}

		rsp := Result{
			IP:     value,
			Result: make(map[string]string),
		}

		for k, v := range resMap {
			rsp.Result[k] = string(v.([]byte))
		}

		data, err := json.Marshal(rsp)
		if err != nil {
			ErrRsp(w, err)
			return
		}

		w.Write(data)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
