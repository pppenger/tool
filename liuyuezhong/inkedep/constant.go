package main

var baseModulePath = map[string]bool{
	"git.inke.cn/inkelogic":       true,
	"git.inke.cn/BackendPlatform": true,
}

var baseModulePath2 = map[string]bool{
	"git.inke.cn/tpc/inf/ikio":             true,
	"git.inke.cn/tpc/inf/metrics":          true,
	"git.inke.cn/tpc/inf/go-upstream":      true,
	"git.inke.cn/tpc/inf/link.base.common": true,
}

// use master branch by default
var necessaryDeps = map[string]bool{
	"git.inke.cn/inkelogic/rpc-go":                 true,
	"git.inke.cn/BackendPlatform/golang":           true,
	"git.inke.cn/BackendPlatform/jaeger-client-go": true,
	"git.inke.cn/tpc/inf/go-upstream":              true,
	"git.inke.cn/inkelogic/daenerys":               true,
	"git.inke.cn/tpc/inf/metrics":                  true,
}

// 灰度配置
type grayConfig struct {
	Jobs []string `toml:"jobs"`

	Exclude []string `toml:"exclude"`

	ServiceTree []struct {
		Owt          string `toml:"owt"`
		Pdl          string `toml:"pdl"`
		ServiceGroup string `toml:"servicegroup"`
		Cluster      string `toml:"cluster"`
	} `toml:"node"`

	Cluster struct {
		Open bool     `toml:"open"`
		List []string `toml:"clusters"`
	} `toml:"cluster"`

	Customs map[string]struct {
		Dep string `toml:"dep"`
		Tag string `toml:"tag"`
		Rev string `toml:"rev"`
	} `toml:"customs"`

	DDRobot struct {
		Open bool   `toml:"open"`
		Url  string `toml:"url"`
	} `toml:"robot"`

	GrayLog string `toml:"gray_log"`
}

var gConfig = &grayConfig{}

var buildMachine = false

type DepsMeta struct {
	Service  string `json:"service"`
	Cluster  string `json:"cluster"`
	Dae      string `json:"dae"`
	RpcGo    string `json:"rpcgo"`
	Upstream string `json:"upstream"`
	Golang   string `json:"golang"`
	Metric   string `json:"metrics"`
	Jaeger   string `json:"jaeger"`
}
