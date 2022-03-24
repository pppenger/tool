# 代码生成工具使用手册
项目模板生成工具

## 目录

* [安装前准备](#安装前准备)

* [开始安装](#开始安装)

* [生成一个 HTTP 服务](#生成一个HTTP服务)
  * [创建新的项目](#创建新的项目http)
  * [目录结构介绍](#目录结构介绍http)
  * [需要注意的事项](#需要注意的事项http)
* [生成一个 RPC 服务](#生成一个RPC服务)
  * [创建新的项目](#创建新的项目)
  * [目录结构介绍](#目录结构介绍)
* [框架使用必读](#框架使用必读)

## 安装前准备

代码自动生成工具的安装是依赖基础库的。同安装其他 golang 库一样，您需要提前安装 Golang 并配置好 Golang 的开发环境。同时需要保证基础库代码是最新的。下面是一些 checklist，请确认本地环境。

1.  首先需要安装[Go](https://golang.org/dl/) (**version 1.10+ is required**)

2.  配置好 GOPATH 路径，GOPATH 的介绍可以参考文档[GOPATH](https://github.com/golang/go/wiki/GOPATH)

3.  基础库代码保持最新。

基础库代码更新步骤如下：

* 由于工具包的安装依赖 daenerys 包, 下面我们来下载 daenerys 包。

> 如果本地没有基础库 daenerys 包，那么可以通过 inkedep get 或者 go get 下载。

```shell
$ inkedep get git.inke.cn/inkelogic/daenerys
或：
$ go get -v -u git.inke.cn/inkelogic/daenerys

```




* 代码生成工具同时也依赖与更新基础库的 golang 包,保持[**master**]()最新代码。
  > 如果本地没有基础库 golang 包，那么可以通过 inkedep get 或者 go get 下载。

```shell
$ inkedep get git.inke.cn/BackendPlatform/golang
或：
$ go get -v -u git.inke.cn/BackendPlatform/golang
```



## 开始安装

> 下载代码生成工具包, 使用 master 分支最新代码。

```shell
$ go get -v -u git.inke.cn/BackendPlatform/daenerys-tool/...
```



> 如果安装成功，则可以直接运行 daenerys 命令查看一些提示信息。

```shell
$ daenerys

A Fast and Flexible Static Site Generator built with
love by spf13 and friends.
Complete documentation is available at http://hugo.spf13.com

Usage:
  daenerys [flags]
  daenerys [command]

Available Commands:
  help        Help about any command
  new         Create a new Daenerys project.
  tool        Daenerys is a very fast static site generator
  version     Print the version number of Daenerys

Flags:
  -h, --help   help for daenerys

Use "daenerys [command] --help" for more information about a command.
```

> 部分同学在安装过程中可能遇到找不到 proto 文件的错误。那么需要更新 protoc 工具,建议使用[**3.6.1**]()以上版本。

```shell
$ brew install protobuf

$ protoc --version
libprotoc 3.6.1
```

## 生成一个 HTTP 服务

### 创建新的项目

在开始生成项目之前，请确认代码生成工具是否安装成功。

运行下面的命令来生成一个 HTTP 项目。

```shell
$ daenerys new --type http <项目名>
或者（快捷方式）：
dae new <项目名>
```


* 测试服务状态

> 命令执行完成后，会在当前路径下生成一个项目名的目录，例如：demo。

> 进入到 demo/app 下，执行 go build 编译该项目，编译成功后会在当前路径下生成可执行文件。直接运行改文件即可启动服务。toml 配置文件中配置了服务的监听端口为 10000。待服务启动后，直接通过 curl 命令或者浏览器访问即可。

```shell
$ cd demo/app
$ go build
$ ./app --config ./config/ali-test/config.toml
```

使用 curl 命令访问服务, 来测试服务是否运行正常。

```shell
$ curl 'http://localhost:10000/ping'

{
    "dm_error": 0,
    "error_msg": "0",
    "data": {
        "result": "ok"
    }
}

$
```

## 目录结构介绍

http server 服务的目录结构如下：

```text
 .
 ├── README.md
 ├── api
 │   └── demo
 │       └── demo.proto // proto文件
 ├── app
 │   ├── build.sh // 编译脚本,发布系统使用
 │   ├── config
 │   │   └── ali-test
 │   │       └── config.toml // 服务的配置文件
 │   └── main.go // main入口
 ├── conf
 │   └── config.go // 服务自身的配置结构
 ├── dao
 │   └── dao.go // DB相关
 ├── manager
 │   └── manager.go // 一些外部接口的封装:rpc client, http client, 中间件等
 ├── model
 │   └── model.go // 业务数据结构模型
 ├── server
 │   └── http
 │       ├── handler.go // 可以将路由的处理逻辑都放在这里,或根据需要写到与handler同级的单独文件中
 │       └── http.go // http服务, 在此处注册路由信息，pb方式会自动注册，默认会注册ping路由
 └── service
     └── service.go // 具体业务逻辑实现
```

> 服务启动后生成的一些日志文件说明:
>
> 详细日志介绍请见：[日志](http://godoc.inkept.cn/logging.html)
```text
* logs // 服务日志目录
* balance日志:服务发现相关日志
* gen日志:框架日志
* stat日志:统计日志，告警监控相关
* business日志:该服务作为client请求其他模块的日志
* access日志:该服务作为server,其他模块请求该服务的日志
* debug/info/error日志:业务逻辑日志
* stdout/stderr日志:其他日志或崩溃日志
* crash日志:服务崩溃日志
```

### 需要注意的事项

服务发布需要生成 Godeps 依赖文件，这样才能将服务正确部署到相应环境上。

> 请确认本地环境已安装了最新的[inkedep](https://wiki.inkept.cn/display/INKE/inkedep-v2)工具。

```shell
// 生成依赖文件
cd app
inkedep save
```

## 生成一个 RPC 服务

### 创建新的项目

运行下面的命令来生成一个 RPC 项目,会在当前运行路径下生成。


> 示例：demo.proto

```proto
syntax = "proto2";

package demo;

import "git.inke.cn/BackendPlatform/daenerys-tool/protoc-gen-daenerys/http/annotations/http.proto";

service Messaging {
    rpc GetMessage (GetMessageReq) returns (GetMessageReply) {
        option (daenerys.api.http) = {
            pattern:"/v1/messages"  // http 路由
            method: "get"           // http 方法
        };
    }
}

// http 请求将会解析到改结构中
message GetMessageReq {
    optional string message_id = 1; // mapped to the URL
    required string name = 2;
}

message GetMessageReply {
    optional string text = 1; // content of the resource
}

```



```shell
$ dae new --type rpc --proto demo.proto rpcdemo
```

编译并运行服务

```shell
$ cd demo/app
$ go build
$ ./app --config ./config/ali-test/config.toml
```

## 目录结构介绍

```
.
├── README.md
├── api
│   └── demo.pb.go
├── app
│   ├── build.sh
│   ├── config
│   │   └── ali-test
│   │       └── config.toml
│   └── main.go
├── conf
│   └── config.go
├── dao
│   └── dao.go
├── manager
│   └── manager.go
├── model
│   └── model.go
├── server
│   └── rpc
│       └── rpc.go
└── service
    └── service.go
```

### 框架使用必读

* [框架使用文档](http://godoc.inkept.cn/)
* [映客代码规范](https://git.inke.cn/BackendPlatform/go-guide)

### 其他推荐
* [The Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
* [Trying Clean Architecture on Golang](https://hackernoon.com/golang-clean-archithecture-efd6d7c43047)
* [google.api.HttpRule](https://cloud.google.com/endpoints/docs/grpc-service-config/reference/rpc/google.api#google.api.HttpRule)

---

Development supported by [INF](http://wiki.inkept.cn/pages/viewpage.action?pageId=1639983).
