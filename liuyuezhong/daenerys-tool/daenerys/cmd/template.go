package cmd

var (
	_tplRPCReadme = `# {{.Name}}
## 项目简介
## 配置说明
## CHANGELOG
`

	_tplRPCMain = `package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"{{.Path}}/conf"
	"{{.Path}}/server/rpc"
	"{{.Path}}/service"
	"git.inke.cn/BackendPlatform/golang/logging"
	"git.inke.cn/inkelogic/daenerys"
)

func init() {
	configS := flag.String("config", "config/config.toml", "Configuration file")
	appS := flag.String("app", "", "App dir")
	flag.Parse()
	
	daenerys.Init(
		daenerys.ConfigPath(*configS),
	)
	
	if *appS != "" {
		daenerys.InitNamespace(*appS)
	}

}

func main() {
	log.Println("{{.Name}} start")

	defer daenerys.Shutdown()

	// init local config
	cfg, err := conf.Init()
	if err != nil {
		logging.Fatalf("service config init error %s", err)
	}

	// create a service instance
	svc := service.New(cfg)

	// init and start http server
	rpc.Init(svc, cfg)

	defer rpc.Shutdown()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sigChan
		log.Printf("get a signal %s\n", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("{{.Name}} server exit now...")
			return
		case syscall.SIGHUP:
		default:
		}
	}

}`

	_tplRPCConf = `package conf

import (
	"git.inke.cn/inkelogic/daenerys"
)

type Config struct {
}

func Init() (*Config, error) {
	// parse Config from config file
	cfg := &Config{}
	err := daenerys.ConfigInstance().Scan(cfg)
	return cfg, err 
}`

	_tplRPCDao = `package dao

import (
	"context"

	"{{.Path}}/conf"
)

// Dao represents data access object
type Dao struct {
	c *conf.Config
}

func New(c *conf.Config) *Dao {
	return &Dao{
		c: c,
	}
}

// Ping check db resource status
func (d *Dao) Ping(ctx context.Context) error {
	return nil
}

// Close release resource
func (d *Dao) Close() error {
	return nil
}

`

	_tplRPCManager = `package manager

import (
	"context"

	"{{.Path}}/conf"
)

// Manager represents middleware component
// such as, kafka, http client or rpc client, etc.
type Manager struct {
	c *conf.Config
}

func New(conf *conf.Config) *Manager {
	return &Manager{
		c: conf,
	}
}


// Ping check middleware resource status
func (m *Manager) Ping(ctx context.Context) error {
	return nil
}

// Close release resource
func (m *Manager) Close() error {
	return nil
}

`

	_tplRPCModel = `package model

type Model struct {

}
`

	_tplRPCServer = `package rpc

import (
	api "{{.Path}}/api/{{.Package}}"
	"{{.Path}}/conf"
	"{{.Path}}/service"
	"git.inke.cn/BackendPlatform/golang/logging"
	"git.inke.cn/inkelogic/daenerys"
	rpcserver "git.inke.cn/inkelogic/daenerys/rpc/server"
	rpcplugin "git.inke.cn/inkelogic/daenerys/plugins/rpc"
)


var (
	svc *service.Service

	rpcServer rpcserver.Server
)

// Init create a rpc server and run it
func Init(s *service.Service, conf *conf.Config) {
	svc = s

	rpcServer = daenerys.RPCServer()

	// add namespace plugin
	rpcServer.Use(rpcplugin.Namespace)

	api.Register{{.ServiceName}}Handler(rpcServer, svc)

	// start a rpc server
	if err := rpcServer.Start(); err != nil {
		logging.Fatalf("rpc server start failed, err %v", err)
	}
}

// Close close the resource
func Shutdown() {
	if rpcServer != nil {
		rpcServer.Stop()
	}
	if svc != nil {
		svc.Close()
	}
}

`

	_tplRPCAppToml = `name="{{.Name}}"
[server]
	service_name="{{.Name}}"
	port = 10000

[log]
	level="debug"
	logpath="logs"
	rotate="hour"

`

	_tplRPCService = `package service

import (
	"context"
	api "{{.Path}}/api/{{.Package}}"
	"{{.Path}}/manager"
	"{{.Path}}/conf"
	"{{.Path}}/dao"
)

// Service represents several business logic(s)
type Service struct {
	c *conf.Config

	// dao: db handler
	dao *dao.Dao

	// manager: other client(s), other middleware(s)
	mgr *manager.Manager
}

// New new a service and return.
func New(c *conf.Config) (s *Service) {
	return &Service{
		c:   c,
		dao: dao.New(c),
		mgr: manager.New(c),
	}
}

// Ping check service's resource status
func (s *Service) Ping(ctx context.Context) error {
	return s.dao.Ping(ctx)
}

// Close close the resource
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
	if s.mgr != nil {
		s.mgr.Close()
	}
}

{{range .Methods}}
func (s *Service) {{.Name}}(ctx context.Context, request *api.{{.InputType}}) (*api.{{.OutputType}}, error) {
	return nil, nil
}
{{end}}

`

	_tplBuild = `#!/bin/bash

cluster_name=$(echo "$1" | sed -r 's/^cop\.([^_\.]+)?_owt\.([^_\.]+)?_pdl\.([^_\.]+)?_cluster\.([^_\.]+)?.*/\4/')
servicegroup_name=$(echo "$1" | sed -r 's/^cop\.([^_\.]+)?_owt\.([^_\.]+)?_pdl\.([^_\.]+)?(.*)?_servicegroup\.([^_\.]+)?.*/\5/')
service_name=$(echo "$1" | sed -r 's/^cop\.([^_\.]+)?_owt\.([^_\.]+)?_pdl\.([^_\.]+)?(.*)?_service\.([^_\.]+)?.*/\5/')
job_name=$(echo "$1" | sed -r 's/^cop\.([^_\.]+)?_owt\.([^_\.]+)?_pdl\.([^_\.]+)?(.*)?_job\.([^_\.]+)?.*/\5/')

#binary name
target=$2
default_target=service

#cluster name
cluster=${cluster_name##*.}

#project path
project_path=$(cd $(dirname $0); pwd)
#src path
src_path=${project_path}
#release path
release_path=release
#bin path
release_bin_path=${release_path}/bin/
#config path
release_config_path=${release_path}/config/


if [ -d "src" ]; then
    printf "find src directory，use src directory \n"
    src_path=${project_path}/src
fi

if [ -d "app" ]; then
    printf "find app directory，use app directory \n"
    src_path=${project_path}
fi

if [ ! $target ]; then
    target=${default_target}
    printf "target is null,use default target name,%s \n" $target
fi
printEnv(){
    printf "Print Env \n"
    printf "============================================\n"
    printf "Commond Params        | %s %s \n" $1  $2
    printf "Project Path          | %s\n" $project_path
    printf "Src Path              | %s\n" $src_path
    printf "Target                | %s\n" $target
    printf "Service Nmae          | %s\n" $service_name
    printf "Cluster Name          | %s\n" $cluster_name
    printf "Cluster 			  | %s\n" $cluster
    printf "Job Name              | %s\n" $job_name
    printf "Service Group Name    | %s\n" $servicegroup_name
	
    printf "Release Path          | %s\n" $release_path
    printf "Release Bin  Path     | %s\n" $release_bin_path
    printf "Release Config Path   | %s\n" $release_config_path
    printf "============================================\n\n\n"
}

cleanDir(){
    printf "Clean Release Dir \n"
    printf "============================================\n"
    cd $project_path
    rm -rf $release_path
    if [ $? != 0 ]; then
        printf "Clean release dir failed\n"
        exit 101
    else
        printf "Clean release dir successed\n"
    fi

    mkdir -p $release_config_path
    mkdir -p $release_bin_path
    printf "============================================\n\n\n"
}

buildBin(){
    printf "Build Bin \n"
    printf "============================================\n"
    cd $src_path
    printf "Pull dependence  ...\n"
    inkedep build
    if [ $? != 0 ]; then
        printf "Compiling project failed\n"
        exit 100
    fi
    printf "Pull dependence End\n"
    printf "Compiling project ...\n"

    go build -o $project_path/release/bin/$target
    if [ $? != 0 ]; then
        printf "Compiling project failed\n"
        exit 102
    else
	    printf "Compiling project successed\n"
    fi
    cd $project_path
    printf "============================================\n\n\n"
}

copyConf(){
    printf "Copy Conf Files\n"
    printf "============================================\n"
    cd $project_path
    cp -r config/$cluster/* release/config/
    echo "Copying config/$cluster into release/config"
    if [ $? != 0 ]; then
        printf "Copying conf failed\n"
        exit 103
    fi
    printf "============================================\n\n\n"
}
	
printRelease(){
    printf "Print Release Directory\n"
    printf "============================================\n"
    cd $project_path
    find $release_path
    printf "============================================\n\n\n"
}
printEnv
cleanDir
buildBin
copyConf
printRelease
exit 0
}
`
)
