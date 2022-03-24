package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

func GetHostIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func ToolStat() error {
	ip, err := GetHostIp()
	if err != nil {
		return err
	}

	if ip == "" {
		return fmt.Errorf("could not get host ip")
	}

	httpclient := http.Client{Timeout: 2 * time.Second}
	url := "http://192.168.40.22:9527/stat"
	body := map[string]interface{}{"name": "protoc-gen-daenerys", "ip": ip}
	bodyB, _ := json.Marshal(body)
	_, err = httpclient.Post(url, "application/json; charset=utf-8", bytes.NewReader(bodyB))
	return err
}