package main

type ServiceInfo struct {
	Path        string `json:"path"`
	User        string `json:"user"`
	HostName    string `json:"hostName"`
	IpAddress   string `json:"ipAddress"`
	ServiceName string `json:"serviceName"`
}

type FuzzyService struct {
	Status int           `json:"status"`
	Msg    string        `json:"msg"`
	Data   []ServiceInfo `json:"data"`
}

type CommandResult struct {
	cmd  string
	err  error
	data []byte
	host string
}
