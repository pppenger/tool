package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// QueryService 查询服务名所在的机器信息
func QueryService(srv string) ([]ServiceInfo, error) {
	u, err := url.Parse(config.SrvUrl)
	if err != nil {
		return nil, fmt.Errorf("config srv_url is invalid: %s", config.SrvUrl)
	}

	u.Path = "/api/v1/query/fuzzy/service/" + srv

	if verbose {
		fmt.Println("http query: ", u.String())
	}

	rsp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error on query service: %w", err)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("error on read query respond: %w", err)
	}

	var fs FuzzyService
	if err := json.Unmarshal(data, &fs); err != nil {
		return nil, fmt.Errorf("error on json Unmarshal: %w", err)
	}

	if fs.Status != 0 {
		return nil, fmt.Errorf("error FuzzyService msg: %s", fs.Msg)
	}

	return fs.Data, nil
}
