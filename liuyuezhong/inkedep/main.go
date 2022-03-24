package main

import (
	"C"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

var usageTemplate = `
inkedep 是公司内部用来管理golang项目的依赖的工具.
更详细的使用方法请参考使用文档:
	https://wiki.inkept.cn/display/INKE/inkedep-v2

使用方法:
	inkedep <command> [arguments]

Commands:
{{range .}}
    {{.Name | printf "%-8s"}} {{.Short}}{{end}}
`

// 全局配置
type repoConfig struct {
	RepoExchanged map[string]struct {
		Old    string `toml:"old"`
		New    string `toml:"new"`
		Branch string `toml:"branch"`
	} `toml:"repoExchanged"`

	BaseModule map[string]string `toml:"baseModule"`
}

var rConfig = &repoConfig{}

// 特殊配置
var configFile = "/build/dependency/config.toml"

func main() {
	flag.Parse()
	args := flag.Args()
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("inkedep: ")

	if len(args) < 1 {
		usageExit()
	}

	// repo地址转换配置
	toml.DecodeFile(configFile, rConfig)

	err := Run(args[0], args[1:])
	if err != nil {
		doPrint("running inkedep command %s error: %s\n", args[0], err)
		os.Exit(-1)
	}
}

func usageExit() {
	tmpl(os.Stdout, usageTemplate, commands)
	os.Exit(-1)
}

func doPrint(fmt string, texts ...interface{}) {
	log.Printf(fmt, texts...)
}

func doWarning(fmts string, texts ...interface{}) {
	doPrint(fmt.Sprintf("[WARNING] %s", fmts), texts...)
}

func doError(fmts string, texts ...interface{}) {
	doPrint(fmt.Sprintf("[ERROR] %s", fmts), texts...)
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("Usage")
	template.Must(t.Parse(strings.TrimSpace(text) + "\n\n\n"))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}
