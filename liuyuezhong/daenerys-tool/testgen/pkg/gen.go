package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"errors"
)

func GenTest(parses []*Parse) (err error) {
	for _, p := range parses {
		if InitMain != "" {
			err = p.genOldTestMain(InitMain)
			return
		}
		switch {
		case strings.HasSuffix(p.Path, "/dao.go") ||
			strings.HasSuffix(p.Path, "/service.go") ||
			strings.HasSuffix(p.Path, "/manager.go"):
			err = p.genTestMain()
		default:
			err = p.genUTTest()
		}
		if err != nil {
			break
		}
	}
	return
}

func (p *Parse) genUTTest() (err error) {
	var (
		buffer  bytes.Buffer
		imports = []string{
			`"context"`,
			`"testing"`,
			`. "github.com/smartystreets/goconvey/convey"`,
		}

		content []byte
	)
	filename := p.Path
	idx := strings.LastIndex(filename, ".go")
	if idx < 0 {
		return errors.New("not a go files")
	}
	filename = filename[:idx] + strings.Replace(filename[idx:], ".go", "_test.go", 1)

	if _, err = os.Stat(filename); (Fn == "" && err == nil) ||
		(err != nil && os.IsExist(err)) {
		err = nil
		return
	}
	for _, impt := range p.Imports {
		if strings.Count(impt.V, "/") > 2 {
			imports = append(imports, fmt.Sprintf(`"%s"`, impt.V))
		}
	}
	imports = RmDup(imports)
	impts := strings.Join(imports, "\n\t")

	if Fn == "" {
		buffer.WriteString(fmt.Sprintf(tpPackage, p.Package))
		buffer.WriteString(fmt.Sprintf(tpImport, impts))
	}
	for _, parseFunc := range p.Funcs {
		if Fn != "" && Fn != parseFunc.Name {
			continue
		}
		var (
			methodK string
			tpVars  string
			vars    []string
			val     []string
			notice  = "Then "
			reset   string
		)
		method := ConvertMethod(p.Path)
		// _, ok := globalImports["rpc-go"]

		if method != "" && !OldStyle {
			methodK = method + "."
		}
		tpTestFuncs := fmt.Sprintf(tpTestFunc, strings.Title(p.Package), parseFunc.Name, "", parseFunc.Name, "%s", "%s", "%s")
		tpTestFuncBeCall := methodK + parseFunc.Name + "(%s)\n\t\t\tConvey(\"%s\", func() {"
		if parseFunc.Result == nil {
			tpTestFuncBeCall = fmt.Sprintf(tpTestFuncBeCall, "%s", "No return values")
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", tpTestFuncBeCall, "%s")
		}
		for k, res := range parseFunc.Result {
			if res.K == "" {
				res.K = fmt.Sprintf("p%d", k+1)
			}
			var so string
			if res.V == "error" {
				res.K = "err"
				so = fmt.Sprintf("\tSo(%s, ShouldBeNil)", res.K)
				notice += "err should be nil."
			} else {
				so = fmt.Sprintf("\tSo(%s, ShouldNotBeNil)", res.K)
				val = append(val, res.K)
			}
			if len(parseFunc.Result) <= k+1 {
				if len(val) != 0 {
					notice += strings.Join(val, ",") + " should not be nil."
				}
				tpTestFuncBeCall = fmt.Sprintf(tpTestFuncBeCall, "%s", notice)
				res.K += " := " + tpTestFuncBeCall
			} else {
				res.K += ", %s"
			}
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", res.K+"\n\t\t\t%s", "%s")
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", "%s", so, "%s")
		}
		if parseFunc.Params == nil {
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", "", "%s")
		}
		for k, pType := range parseFunc.Params {
			if pType.K == "" {
				pType.K = fmt.Sprintf("a%d", k+1)
			}
			var (
				init   string
				params = pType.K
			)
			switch {
			case strings.HasPrefix(pType.V, "context"):
				init = params + " = context.Background()"
			case strings.HasPrefix(pType.V, "[]byte"):
				init = params + " = " + pType.V + "(\"\")"
			case strings.HasPrefix(pType.V, "[]"):
				init = params + " = " + pType.V + "{}"
			case strings.HasPrefix(pType.V, "int") ||
				strings.HasPrefix(pType.V, "uint") ||
				strings.HasPrefix(pType.V, "float") ||
				strings.HasPrefix(pType.V, "double"):
				init = params + " = " + pType.V + "(0)"
			case strings.HasPrefix(pType.V, "string"):
				init = params + " = \"\""
			case strings.Contains(pType.V, "*xsql.Tx"):
				init = params + ",_ = " + methodK + "BeginTran(c)"
				reset += "\n\t" + params + ".Commit()"
			case strings.HasPrefix(pType.V, "*"):
				init = params + " = " + strings.Replace(pType.V, "*", "&", -1) + "{}"
			case strings.Contains(pType.V, "chan"):
				init = params + " = " + pType.V
			case pType.V == "time.Time":
				init = params + " = time.Now()"
			case strings.Contains(pType.V, "chan"):
				init = params + " = " + pType.V
			default:
				init = params + " " + pType.V
			}
			vars = append(vars, "\t\t"+init)
			if len(parseFunc.Params) > k+1 {
				params += ", %s"
			}
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", params, "%s")
		}
		if len(vars) > 0 {
			tpVars = fmt.Sprintf(tpVar, strings.Join(vars, "\n\t"))
		}
		tpTestFuncs = fmt.Sprintf(tpTestFuncs, tpVars, "%s")
		if reset != "" {
			tpTestResets := fmt.Sprintf(tpTestReset, reset)
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, tpTestResets)
		} else {
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "")
		}
		buffer.WriteString(tpTestFuncs)
	}
	var (
		file *os.File
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	)
	if file, err = os.OpenFile(filename, flag, 0644); err != nil {
		return
	}
	if Fn == "" {
		content, _ = GoImport(filename, buffer.Bytes())
	} else {
		content = buffer.Bytes()
	}
	if _, err = file.Write(content); err != nil {
		return
	}
	if err = file.Close(); err != nil {
		return
	}
	return
}

func (p *Parse) genTestMain() (err error) {
	var (
		buffer             bytes.Buffer
		impts              []string
		vars, mainFunc     string
		content            []byte
		instance, confFunc string
		tomlPath           = "./config/ali-test/config.toml"
		filename           = strings.Replace(p.Path, ".go", "_test.go", -1)
	)
	instance = ConvertMethod(p.Path)
	if instance == "s" {
		vars = strings.Join([]string{"s *Service"}, "\n\t")
	} else if instance == "d" {
		vars = strings.Join([]string{"d *Dao"}, "\n\t")
	} else {
		vars = strings.Join([]string{"m *Manager"}, "\n\t")
	}
	mainFunc = tpTestMain
	impts = []string{`"os"`, `"testing"`, `"git.inke.cn/inkelogic/daenerys"`}

	// 添加所有引入包，然后再goimports优化
	for _, v := range globalImports {
		if strings.HasSuffix(v.V, "conf") {
			impts = append(impts, fmt.Sprintf(`"%s"`, v.V))
			break
		}
	}
	imports := strings.Join(impts, "\n\t")
	confFunc = instance + " = New(conf)"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		buffer.WriteString(fmt.Sprintf(tpPackage, p.Package))
		buffer.WriteString(fmt.Sprintf(tpImport, imports))
		buffer.WriteString(fmt.Sprintf(tpVar, vars))
		buffer.WriteString(fmt.Sprintf(mainFunc, tomlPath, tplMock, confFunc))
		content, _ = GoImport(filename, buffer.Bytes())
		ioutil.WriteFile(filename, content, 0644)
	}
	return
}
func (p *Parse) genOldTestMain(tp string) (err error) {
	var (
		buffer   bytes.Buffer
		impts    string
		mainFunc string
		content  []byte
		tomlPath = "./config/ali-test/config.toml"
		filename = "main_test.go"
	)
	mainFunc = tpOldTestMain
	impts = strings.Join([]string{`"os"`, `"testing"`, `"git.inke.cn/inkelogic/rpc-go"`}, "\n\t")
	initFunc := "rpc.NewServerWithConfig(conf)"
	if tp == "http" {
		initFunc = "rpc.NewHTTPServerWithConfig(conf)"
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		buffer.WriteString(fmt.Sprintf(tpPackage, p.Package))
		buffer.WriteString(fmt.Sprintf(tpImport, impts))
		buffer.WriteString(fmt.Sprintf(mainFunc, tomlPath, tplMock, initFunc))
		content, _ = GoImport(filename, buffer.Bytes())
		ioutil.WriteFile(filename, content, 0644)
	}
	fmt.Printf("Generate %s success\n", filename)
	return
}
