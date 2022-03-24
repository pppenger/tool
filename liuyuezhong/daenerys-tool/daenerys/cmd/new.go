package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"go/importer"
	"io"
	"io/ioutil"
	"os"
	path "path/filepath"
	"strings"
	"text/template"

	"git.inke.cn/BackendPlatform/daenerys-tool/daenerys/internal/goparser"
	"git.inke.cn/BackendPlatform/daenerys-tool/daenerys/internal/models"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/spf13/cobra"
)

var (
	p = new(project)
)

type method struct {
	Name       string
	InputType  string
	OutputType string
}

type project struct {
	ServiceName string
	Path        string
	Proto       string
	Dir         string
	Name        string
	Methods     []method
	Type        string
	Router      string
	Handler     string
	Package     string
	Doc         string
}

func init() {
	pwd, _ := os.Getwd()
	newCmd.Flags().StringVar(&p.Proto, "proto", "", "whether to use protobuf to create rpc project")
	newCmd.Flags().StringVar(&p.Dir, "project-dir", pwd, "project directory to create project")
	newCmd.Flags().StringVar(&p.Type, "type", "http", "http or rpc, default http")
	rootCmd.AddCommand(newCmd)
}

const (
	_tplTypeService = iota
	_tplTypeBuild
	_tplTypeManager
	_tplTypeServer
	_tplTypeModel
	_tplTypeReadme
	_tplTypeMain
	_tplTypeAppToml
	_tplTypeDao
	_tplTypeConfig

	_tplTypeHTTPERRCode
	_tplTypeHTTPMain
	_tplTypeHTTPConfig
	_tplTypeHTTPServer
	_tplTypeHTTPService
	_tplTypeGit
	_tplTypeHTTPHandler
	_tplTypeHTTPDao
	_tplTypeHTTPManager
	_tplTypeHTTPModel
	_tplTypePBHTTPService
	_tplTypePBHTTPHandler
	_tplTypeHTTPRouter
	_tplTypeDoc
)

type file struct {
	Name string
	F    func(tpl int) ([]byte, error)
	T    string
}

var (
	files = map[int]file{
		_tplTypeHTTPERRCode: {Name: "api/code/code.go", T: _tplERRCode},
		_tplTypeMain:        {Name: "app/main.go", T: _tplRPCMain},
		_tplTypeBuild:       {Name: "app/build.sh", T: _tplBuild},
		_tplTypeAppToml:     {Name: "app/config/ali-test/config.toml", T: _tplRPCAppToml},
		_tplTypeConfig:      {Name: "conf/config.go", T: _tplRPCConf},
		_tplTypeDao:         {Name: "dao/dao.go", T: _tplRPCDao},
		_tplTypeManager:     {Name: "manager/manager.go", T: _tplRPCManager},
		_tplTypeModel:       {Name: "model/model.go", T: _tplRPCModel},
		_tplTypeServer:      {Name: "server/rpc/rpc.go", T: _tplRPCServer},
		_tplTypeService:     {Name: "service/service.go", T: _tplRPCService},
		_tplTypeReadme:      {Name: "README.md", T: _tplRPCReadme},
		_tplTypeGit:         {Name: ".gitignore", T: _tplGit},
	}

	hfiles = map[int]file{
		_tplTypeHTTPERRCode:   {Name: "api/code/code.go", T: _tplERRCode},
		_tplTypeHTTPMain:      {Name: "app/main.go", T: _tplHMain},
		_tplTypeBuild:         {Name: "app/build.sh", T: _tplBuild},
		_tplTypeAppToml:       {Name: "app/config/ali-test/config.toml", T: _tplHAppToml},
		_tplTypeHTTPConfig:    {Name: "conf/config.go", T: _tplHConfig},
		_tplTypeHTTPDao:       {Name: "dao/dao.go", T: _tplHDao},
		_tplTypeHTTPManager:   {Name: "manager/manager.go", T: _tplHManager},
		_tplTypeHTTPModel:     {Name: "model/model.go", T: _tplHModel},
		_tplTypeHTTPServer:    {Name: "server/http/http.go", T: _tplHServer},
		_tplTypeHTTPService:   {Name: "service/service.go", T: _tplHService},
		_tplTypeHTTPHandler:   {Name: "server/http/handler.go", T: _tplHHandler},
		_tplTypePBHTTPService: {Name: "service/service.go", T: _tplHPBService},
		_tplTypePBHTTPHandler: {Name: "server/http/handler.go", T: _tplHPBHandler},
		_tplTypeReadme:        {Name: "README.md", T: _tplRPCReadme},
		_tplTypeGit:           {Name: ".gitignore", T: _tplGit},
		_tplTypeHTTPRouter:    {Name: "server/http/router.go", T: _tplHRouter},
		_tplTypeDoc:           {Name: "api/doc.md", T: _tplDoc},
	}
)

var newCmd = &cobra.Command{
	Use:   "new [flags] name",
	Short: "Create a new Daenerys project.",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return errors.New("requires a project name")
		}
		p.Name = args[0]
		tmp, err := path.Rel(path.Join(os.Getenv("GOPATH"), "src"), path.Join(p.Dir, p.Name))
		if err != nil {
			panic(err)
		}
		p.Path = path.ToSlash(tmp)
		return
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.HasPrefix(p.Dir, path.Join(os.Getenv("GOPATH"), "src")) {
			return errors.New("must under the GOPATH/src to create project")
		}
		if !isInstall("protoc-gen-daenerys") {
			install("go", "", nil, "get", "-v", "-u", "git.inke.cn/BackendPlatform/daenerys-tool/protoc-gen-daenerys")
			err := install("go", p.Dir,
				map[string]string{"GOBIN": path.Join(os.Getenv("GOPATH"), "bin")},
				"install", "git.inke.cn/BackendPlatform/daenerys-tool/protoc-gen-daenerys")
			if err != nil {
				return err
			}
		}
		if err := p.create(); err != nil {
			return err
		}
		return nil
	},
}

func (p *project) copy(dst string, src string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		fmt.Printf("read file failed: %s\n", src)
		return
	}
	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		fmt.Printf("copy file failed: %s\n", src)
	}
}

func (p *project) genpb(tp string) error {

	dir := path.Join(p.Dir, p.Name, "api")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpfile, err := ioutil.TempFile("", "protoc-*")
	if err != nil {
		return err
	}
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	deps := path.Join(os.Getenv("GOPATH"), "src")
	outDir := path.Join(p.Dir, p.Name, "api")

	src, _ := path.Abs(p.Proto)
	protoPath := path.Dir(src)
	err = install(
		"protoc", "", nil,
		fmt.Sprintf("--daenerys_out=plugins=%s:%s", tp, outDir),
		"-I="+deps,
		//"--proto_path="+path.Dir(p.Proto),
		"--proto_path="+protoPath,
		"-o", tmpfile.Name(),
		path.Base(p.Proto),
	)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(tmpfile)
	if err != nil {
		return err
	}

	var descSet descriptor.FileDescriptorSet
	if err = proto.Unmarshal(data, &descSet); err != nil {
		return err
	}
	var dst1 string
	for _, desc := range descSet.File {
		for _, service := range desc.Service {
			p.ServiceName = strings.Title(service.GetName())
			for _, mtd := range service.Method {
				p.Methods = append(p.Methods, method{
					Name:       mtd.GetName(),
					InputType:  methodName(mtd.GetInputType()),
					OutputType: methodName(mtd.GetOutputType()),
				})
			}
			break
		}

		pkg := *desc.Package
		p.Package = pkg
		dst1 = path.Join(outDir, pkg)
		if err := os.MkdirAll(dst1, 0755); err != nil {
			return err
		}

		break
	}

	ss := strings.SplitN(p.Proto, ".", -1)
	if len(ss) <= 1 {
		panic("file name invalid")
	}
	//copy pb proto
	dst := path.Join(dst1, p.Proto)
	p.copy(dst, src)
	//copy pb.go
	if p.Type == "rpc" {
		name := strings.Join(ss[:len(ss)-1], ".")
		p.copy(path.Join(dst1, name+".pb.go"), path.Join(outDir, name+".pb.go"))
		os.Remove(path.Join(outDir, name+".pb.go"))
	}
	return nil
}

func (p *project) genservice(tpl int) ([]byte, error) {
	return p.parse(_tplRPCService)
}

func (p *project) parse(s string) ([]byte, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *project) genmock(tpl int) ([]byte, error) {
	return nil, nil
}

func (p *project) gendefault(tpl int) ([]byte, error) {
	if p.Type == "http" {
		return p.parse(hfiles[tpl].T)
	}
	return p.parse(files[tpl].T)
}

func methodName(method string) string {
	return strings.Split(method, ".")[2]
}

func (p *project) create() (err error) {
	if p.Type != "rpc" && p.Type != "http" {
		return fmt.Errorf("invalid type")
	}
	if p.Proto != "" {
		if err = p.genpb(p.Type); err != nil {
			return
		}
	}

	if p.Type == "http" {
		routerFile := path.Join(p.Dir, "/.genRouter.txt")
		d, _ := ioutil.ReadFile(routerFile)
		buf := bytes.NewBuffer(d)
		p.Router = buf.String()
		os.Remove(routerFile)

		handlerFile := path.Join(p.Dir, "/.genHandler.txt")
		d, _ = ioutil.ReadFile(handlerFile)
		buf = bytes.NewBuffer(d)
		p.Handler = buf.String()
		os.Remove(handlerFile)

		docFile := path.Join(p.Dir, "/.doc.md")
		d, _ = ioutil.ReadFile(docFile)
		buf = bytes.NewBuffer(d)
		p.Doc = buf.String()
		os.Remove(docFile)

		withPB := false
		if len(p.Router) > 0 || len(p.Handler) > 0 {
			withPB = true
		}

		for tpl, f := range hfiles {
			dir := path.Dir(path.Join(p.Dir, p.Name, f.Name))
			if err = os.MkdirAll(dir, 0755); err != nil {
				return
			}
			var data []byte
			if f.F != nil {
				data, err = f.F(tpl)
			} else {
				data, err = p.gendefault(tpl)
			}
			if err != nil {
				return
			}

			if !withPB && tpl == _tplTypePBHTTPHandler {
				continue
			}

			if !withPB && tpl == _tplTypePBHTTPService {
				continue
			}

			if withPB && tpl == _tplTypeHTTPService {
				continue
			}

			if withPB && tpl == _tplTypeHTTPHandler {
				continue
			}

			if withPB && tpl == _tplTypeHTTPModel {
				modelFile := path.Join(p.Dir, "/.genModel.txt")
				pbModel, _ := ioutil.ReadFile(modelFile)
				buf := bytes.NewBuffer(pbModel)
				data = append(data, buf.Bytes()...)
				os.Remove(modelFile)
			}
			if withPB && tpl == _tplTypePBHTTPHandler {
				bodyFile := path.Join(p.Dir, "/.genBody.txt")
				body, _ := ioutil.ReadFile(bodyFile)
				buf := bytes.NewBuffer(body)
				data = append(data, buf.Bytes()...)
				os.Remove(bodyFile)
			}
			if withPB && tpl == _tplTypePBHTTPService {
				files, err := ioutil.ReadDir(path.Join(p.Dir, p.Name, "service/"))
				var fileNames []models.Path
				var existedMethods = make(map[string]struct{})
				for _, f := range files {
					if !f.IsDir() {
						fileNames = append(fileNames, models.Path(path.Join(p.Dir, p.Name, "service", f.Name())))
					}
				}
				parser := goparser.Parser{Importer: importer.Default()}
				if err == nil {
					for _, name := range fileNames {
						result, err := parser.Parse(string(name), fileNames)
						if err == nil {
							for _, f := range result.Funcs {
								existedMethods[f.Name] = struct{}{}
							}
						} else {
							return err
						}
					}
				}
				if len(existedMethods) != 0 {
					var newMethod []method
					for _, method := range p.Methods {
						if _, exist := existedMethods[method.Name]; exist {
							continue
						}
						newMethod = append(newMethod, method)
					}
					var d []byte
					if len(newMethod) != 0 {
						p := project{Methods: newMethod}
						d, err = p.parse(_tplMethods)
						if err != nil {
							return err
						}
					}
					err = appendFile(path.Join(p.Dir, p.Name, f.Name), d, 0644)
					if err != nil {
						return err
					}
				}
			}

			err = overwriteGenerateFile(path.Join(p.Dir, p.Name, f.Name), data, 0644)
			if err != nil {
				return
			}
		}
		p.createHTTPPBService()
		return nil
	}

	for tpl, f := range files {
		dir := path.Dir(path.Join(p.Dir, p.Name, f.Name))
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
		var data []byte
		if f.F != nil {
			data, err = f.F(tpl)
		} else {
			data, err = p.gendefault(tpl)
		}
		if err != nil {
			return
		}

		err = ioutil.WriteFile(path.Join(p.Dir, p.Name, f.Name), data, 0644)
		if err != nil {
			return
		}

	}
	return nil
}

func overwriteGenerateFile(filename string, data []byte, perm os.FileMode) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) || bytes.Contains(data, []byte("DO NOT EDIT")) {
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		n, err := f.Write(data)
		if err == nil && n < len(data) {
			err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
			err = err1
		}
		return err
	}
	return nil
}

func appendFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func (p *project) createHTTPPBService() {
	if p.Type == "http" {
		//cur, _ := os.Getwd()
		//headerFile := path.Join(cur, "/.genHeader.txt")
		//apiFile := path.Join(cur, "/.genApi.txt")
		//header, _ := ioutil.ReadFile(headerFile)
		//api, _ := ioutil.ReadFile(apiFile)
		//
		//all := bytes.NewBuffer([]byte{})
		//all.WriteString("package api\n\n")
		//all.Write(header)
		//all.WriteString("import (")
		//all.WriteString(fmt.Sprintf(`"%s/model"`, p.Path))
		//all.WriteString("\n)\n")
		//all.Write(api)
		//
		//resultFile := "api/generated.go"
		//ioutil.WriteFile(path.Join(p.Dir, p.Name, resultFile), all.Bytes(), 0644)
		//os.Remove(headerFile)
		//os.Remove(apiFile)
	}
}
