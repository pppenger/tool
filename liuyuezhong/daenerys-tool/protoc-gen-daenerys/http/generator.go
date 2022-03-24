package http

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"reflect"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"

	"git.inke.cn/BackendPlatform/daenerys-tool/protoc-gen-daenerys/comment"
	"git.inke.cn/BackendPlatform/daenerys-tool/protoc-gen-daenerys/http/annotations"

	"github.com/golang/protobuf/proto"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	clientPkgPath  = "git.inke.cn/inkelogic/daenerys/http/client"
	serverPkgPath  = "git.inke.cn/inkelogic/daenerys/http/server"
)

func init() {
	generator.RegisterPlugin(new(http))
}

// http is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for go-http support.
type http struct {
	gen      *generator.Generator
	resolver *comment.Resolver
}

// Name returns the name of this plugin, "httpmicro".
func (g *http) Name() string {
	return "http"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	clientPkg  string
	serverPkg  string
	modelPkg   string
)

// Exported buffer
var (
	GenBodyBuffer    = bytes.NewBuffer([]byte{})
	GenPBModelBuffer = bytes.NewBuffer([]byte{})
	GenRouteBuffer   = bytes.NewBuffer([]byte{})
	GenHandlerBuffer = bytes.NewBuffer([]byte{})
	DocBuffer        = bytes.NewBuffer([]byte{})
)

// Init initializes the plugin.
func (g *http) Init(gen *generator.Generator) {
	g.gen = gen
	contextPkg = generator.RegisterUniquePackageName("context", nil)
	clientPkg = generator.RegisterUniquePackageName("httpclient", nil)
	serverPkg = generator.RegisterUniquePackageName("httpserver", nil)
	modelPkg = generator.RegisterUniquePackageName("model", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *http) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *http) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *http) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *http) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.resolver = comment.New([]*descriptor.FileDescriptorProto{file.FileDescriptorProto})

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
		g.generateDoc(file.FileDescriptorProto, service)
	}
}

var goPath = os.Getenv("GOPATH")

// RunPath TODO: NEEDS COMMENT INFO
var RunPath string

// GenerateImports generates the import declaration for this file.
func (g *http) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	cur, _ := os.Getwd()
	RunPath, _ = filepath.Rel(filepath.Join(goPath, "src"), cur)
	RunPath = filepath.Join(RunPath, file.GetPackage())
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
	// TODO: do we need any in go-httpmicro?
}

func unexport(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

var pat = `\{([a-zA-Z_.-]+)\}`
var reg = regexp.MustCompile(pat)

func (g *http) generateDoc(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	serviceComment, err := g.resolver.ServiceComments(file, service)
	if err != nil {
		g.gen.Error(err, "获取服务文档失败，proto中必须有服务说明")
		return
	}
	lines := strings.Split(serviceComment.Leading, "\n")
	DocBuffer.WriteString(fmt.Sprintf("## 1.%s\n\n", strings.Join(getCommentNote(nameTag, lines, ""), "\n")))
	DocBuffer.WriteString(fmt.Sprintf("%s\n", strings.Join(getCommentNote(descTag, lines, ""), "\n")))
	g.genMethodDoc(DocBuffer, file, service)

}

func getCommentNote(tag string, comments []string, prefix string) []string {
	notes := make([]string, 0)
	for _, c := range comments {
		cTrimed := strings.TrimSpace(c)
		if strings.HasPrefix(cTrimed, tag) {
			notes = append(notes, prefix+strings.TrimPrefix(cTrimed, tag))
		}
	}
	return notes
}

const (
	nameTag  = "[name]"
	descTag  = "[desc]"
	sceneTag = "[scene]"
	logicTag = "[logic]"
	noteTag  = "[note]"
)

func (g *http) genMethodDoc(DocBuffer *bytes.Buffer, file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {

	for i, method := range service.Method {
		methodComment, err := g.resolver.MethodComments(file, service, method)
		if err != nil {
			g.gen.Error(err, "获取服务文档失败，method中必须有接口说明")
			continue
		}
		httpMethod, matchedPath, body := g.getHttpInfo(service, method)
		lines := strings.Split(methodComment.Leading, "\n")
		DocBuffer.WriteString(fmt.Sprintf("\n### %d. %s\n", i+1, strings.Join(getCommentNote(nameTag, lines, ""), "\n")))
		DocBuffer.WriteString(fmt.Sprintf("1. **接口描述**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("%s\n", strings.Join(getCommentNote(descTag, lines, "    >"), "\n")))
		DocBuffer.WriteString(fmt.Sprintf("\n2. **应用场景**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("%s\n", strings.Join(getCommentNote(sceneTag, lines, "    >"), "\n")))

		DocBuffer.WriteString(fmt.Sprintf("\n3. **重要逻辑**(选填)\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("%s\n", strings.Join(getCommentNote(logicTag, lines, "    >"), "\n")))
		DocBuffer.WriteString(fmt.Sprintf("\n4. **调用方式**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("   > HTTP %s\n", strings.TrimSpace(httpMethod)))
		DocBuffer.WriteString(fmt.Sprintf("\n5. **服务地址**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("   > 生产 https://{host}/%s\n\n", strings.TrimSpace(matchedPath)))
		DocBuffer.WriteString(fmt.Sprintf("   > 测试 https://{host}/%s\n", strings.TrimSpace(matchedPath)))
		DocBuffer.WriteString(fmt.Sprintf("\n6. **sdk地址**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("   > sdk git地址\n"))
		DocBuffer.WriteString(fmt.Sprintf("\n7. **请求参数**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("   |参数名|参数类型|是否必填|示例值|说明|\n"))
		DocBuffer.WriteString(fmt.Sprintf("   |------|--------|--------|------|----|\n"))

		var line = &[]string{}
		g.getParams(file, method.GetInputType(), line, "")
		DocBuffer.WriteString(strings.Join(*line, "\n"))
		DocBuffer.WriteString("\n")
		DocBuffer.WriteString(fmt.Sprintf("\n8. **请求示例**\n\n"))

		md := g.resolver.Message(method.GetInputType())
		var buf = &[]string{}
		g.example(md, file, buf, "", 4, "")
		j := strings.Join(*buf, "    \n")
		DocBuffer.WriteString("    ```json\n")
		DocBuffer.WriteString(j)
		DocBuffer.WriteString("\n")
		DocBuffer.WriteString("    ```\n")

		DocBuffer.WriteString(fmt.Sprintf("\n9. **响应参数**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("   |参数名|参数类型|是否必填|示例值|说明|\n"))
		DocBuffer.WriteString(fmt.Sprintf("   |------|--------|--------|------|----|\n"))
		line = &[]string{}
		g.getParams(file, method.GetOutputType(), line, "")
		DocBuffer.WriteString(strings.Join(*line, "\n"))
		DocBuffer.WriteString("\n")
		DocBuffer.WriteString(fmt.Sprintf("\n10. **响应示例**\n\n"))

		md = g.resolver.Message(method.GetOutputType())
		buf = &[]string{}
		g.example(md, file, buf, "", 8, "")
		j = strings.Join(*buf, "        \n")
		DocBuffer.WriteString("    ```json\n")
		DocBuffer.WriteString("    {\n")
		DocBuffer.WriteString(`        "dm_error": 0,`)
		DocBuffer.WriteString("\n")
		DocBuffer.WriteString(`        "error_msg": "操作成功",`)
		DocBuffer.WriteString("\n")
		DocBuffer.WriteString(`        "data":`)
		DocBuffer.WriteString(j)
		DocBuffer.WriteString("\n    }\n")
		DocBuffer.WriteString("    ```\n")
		DocBuffer.WriteString(fmt.Sprintf("\n11. **注意事项**\n\n"))
		DocBuffer.WriteString(fmt.Sprintf("%s\n", strings.Join(getCommentNote(noteTag, lines, "    >"), "\n")))

		_ = body
	}
}

func (g *http) getParams(file *descriptor.FileDescriptorProto, name string, line *[]string, parent string) {
	md := g.resolver.Message(name)
	for _, f := range md.Descriptor.Field {
		val, tp, builtin := mockFiled(f)
		name := f.GetName()
		fieldComment, _ := g.resolver.FieldComments(file, md, f)
		if parent != "" {
			name = parent + "." + name
		}
		*line = append(*line, fmt.Sprintf("   | %s| %s| %t| %s| %s|", name, strings.TrimPrefix(tp, "."+file.GetPackage()+"."), isRequire(f), val, strings.TrimSpace(strings.Split(fieldComment.Leading, "\n")[0])))
		if !builtin {
			g.getParams(file, f.GetTypeName(), line, name)
		}
	}
}
func indentN(i int) string {
	return strings.Repeat(" ", i)
}
func isBuiltin(field *descriptor.FieldDescriptorProto) bool {
	if field.Type == nil {
		return false
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_BOOL,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_BYTES,
		descriptor.FieldDescriptorProto_TYPE_STRING:
		return true
	default:
		return false
	}
}
func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
}
func isRequire(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED
}
func isOptional(field *descriptor.FieldDescriptorProto) bool {
	return field.Label == nil || field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL
}

func mockFiled(field *descriptor.FieldDescriptorProto) (val string, tp string, builtin bool) {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		val = field.GetDefaultValue()
		tp = "bool"
		builtin = true
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
		val = field.GetDefaultValue()
		tp = "float"
		builtin = true
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_ENUM,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		val = field.GetDefaultValue()
		tp = "int"
		builtin = true
	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		val = strconv.Quote(field.GetDefaultValue())
		tp = "string"
		builtin = true
	default:
		tp = field.GetTypeName()
		builtin = false
	}
	if isRepeated(field) {
		tp = tp + " Array"
	}
	return
}

func (g *http) example(msg *comment.Message, file *descriptor.FileDescriptorProto, buf *[]string, fieldName string, indent int, outEndComma string) {
	if fieldName == "" {
		*buf = append(*buf, indentN(indent)+"{")
	} else {
		*buf = append(*buf, indentN(indent)+fmt.Sprintf(`"%s": {`, fieldName))
	}
	num := len(msg.Descriptor.Field)
	for i, f := range msg.Descriptor.Field {
		isBuiltin := isBuiltin(f)

		endComma := ""
		if i < (num - 1) {
			endComma = ","
		}
		repeated := isRepeated(f)
		if isBuiltin {
			mockVal, _, _ := mockFiled(f)

			if repeated {
				*buf = append(*buf, indentN(indent+4)+`"`+f.GetName()+`": [`)
				*buf = append(*buf, indentN(indent+8)+mockVal)
				*buf = append(*buf, indentN(indent+4)+`]`+endComma)
			} else {
				*buf = append(*buf, indentN(indent+4)+`"`+f.GetName()+`": `+mockVal+endComma)
			}
		} else {
			subMsg := g.resolver.Message(f.GetTypeName())
			if subMsg == nil {
				panic(fmt.Sprintf("%v%v", f.TypeName, f.Type))
			}
			nextIndent := indent + 4
			nextFname := f.GetName()
			if repeated {
				nextIndent = indent + 8
				nextFname = ""
			}
			if repeated {
				*buf = append(*buf, indentN(indent+4)+`"`+f.GetName()+`":[`)
				g.example(subMsg, file, buf, nextFname, nextIndent, "")
				*buf = append(*buf, indentN(indent+4)+`]`+endComma)
			} else {
				g.example(subMsg, file, buf, nextFname, nextIndent, endComma)
			}
		}
	}
	*buf = append(*buf, indentN(indent)+"}"+outEndComma)
}

// generateService generates all the code for the named service.
func (g *http) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, index int) {
	serviceName := strings.ToLower(service.GetName())
	if pkg := file.GetPackage(); pkg != "" {
		serviceName = pkg
	}
	_ = serviceName
	for _, method := range service.Method {
		lowName := unexport(method.GetName())
		inType := g.typeName(method.GetInputType())
		GenHandlerBuffer.WriteString(fmt.Sprintf("\nfunc %s(c *%s.Context) {\n", lowName, serverPkg))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\treq := new(%s.%s)\n", modelPkg, inType))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\tif err := c.Bind(c.Request, req); err != nil {\n"))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\t\tc.JSONAbort(nil, err)\n"))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\t\treturn\n"))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\t}\n"))
		GenHandlerBuffer.WriteString(fmt.Sprintf("\tresp, err := svc.%s(c.Ctx, req)\n", method.GetName()))
		GenHandlerBuffer.WriteString("\tc.JSON(resp, err)\n")
		GenHandlerBuffer.WriteString("}\n\n")
	}
	GenBodyBuffer.WriteString("\n")

	matchedFiled := map[string]bool{}
	outTypeMap := map[string]bool{}
	var body string
	for _, method := range service.Method {
		outType := g.typeName(method.GetOutputType())
		outTypeMap[outType] = true
		lowName := unexport(method.GetName())
		var httpMethod, matchedPath string
		httpMethod, matchedPath, body = g.getHttpInfo(service, method)
		//获取匹配的字段
		r := reg.FindAllStringSubmatch(matchedPath, -1)
		for _, rr := range r {
			if len(rr) > 1 {
				matchedFiled[rr[1]] = true
			}
			//替换参数
			matchedPath = strings.Replace(matchedPath, rr[0], ":"+rr[1], 1)
		}

		//生成路由注册信息
		GenRouteBuffer.WriteString("\t")
		GenRouteBuffer.WriteString(fmt.Sprintf(`s.%s("%s", %s)`, httpMethod, matchedPath, lowName))
		GenRouteBuffer.WriteString("\n")
	}

	var pbModel string
	//生成model数据结构
	for _, m := range file.MessageType {
		pbModel += fmt.Sprintf("type %s struct {\n", *m.Name)
		for _, f := range m.Field {
			fName := generator.CamelCase(*f.Name)

			//字段类型
			fTName := ""
			if f.GetType().String() != "TYPE_MESSAGE" {
				fTName = fieldDescriptorTypeName[f.GetType().String()]
			} else {
				if strings.Contains(f.GetTypeName(), *m.Name) {
					//todo
				}
				ss := strings.Split(f.GetTypeName(), ".")
				if len(ss) > 0 {
					fTName = ss[len(ss)-1]
				}
			}
			if isRepeated(f) {
				fTName = "[]" + fTName
			}

			//判断是否为query参数
			isSchemaTag := true //todo
			for k := range matchedFiled {
				if strings.Contains(k, *f.Name) {
					isSchemaTag = false
					break
				}
				isSchemaTag = true
			}

			//是否设置body,body为*则其他字段都不是query参数
			if strings.Contains(body, *f.Name) || body == "*" {
				isSchemaTag = false
			}

			//out message ignored schema
			if _, ok := outTypeMap[*m.Name]; ok {
				isSchemaTag = false
			}

			validataTag := ` validate:"required" `
			jsonTag := "`json:"
			schemaTag := "schema:"
			l := fmt.Sprintf("\t%s\t%s\t", fName, fTName)
			tag := fmt.Sprintf("%s", jsonTag) + fmt.Sprintf(`"%s" `, f.GetName())
			if isSchemaTag {
				tag += fmt.Sprintf("%s", schemaTag) + fmt.Sprintf(`"%s" `, f.GetName())
			}
			if f.Label.String() == "LABEL_REQUIRED" {
				tag += validataTag
			}
			tag = strings.TrimSpace(tag)
			tag += "`"
			pbModel += fmt.Sprintf(l + tag + "\n")
		}
		pbModel += fmt.Sprintf("}\n\n")
	}

	if len(pbModel) > 0 {
		GenPBModelBuffer.WriteString(pbModel)
	}
}

func (g *http) getHttpInfo(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) (httpMethod string, newPath string, body string) {
	googleOptionInfo, err := ParseMethod(method)
	if err == nil {
		httpMethod = strings.ToUpper(googleOptionInfo.Method)
		p := googleOptionInfo.PathPattern
		body = googleOptionInfo.HTTPRule.GetBody()
		//panic(fmt.Sprintf("%s %s %s", httpMethod, p, body))
		if p != "" {
			newPath = p
			return
		}
	}

	if httpMethod == "" {
		// resolve http method
		httpMethod = getTagValue("method", nil)
		if httpMethod == "" {
			httpMethod = "GET"
		} else {
			httpMethod = strings.ToUpper(httpMethod)
		}
	}
	return
}

// generateServerSignature returns the server-side signature for a method.
func (g *http) generateServerSignature(servName string, method *descriptor.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}

	var reqArgs []string
	rets := []string{}
	reqArgs = append(reqArgs, contextPkg+".Context")

	if !method.GetClientStreaming() {
		reqArgs = append(reqArgs, "*"+modelPkg+"."+g.typeName(method.GetInputType()))
	}
	if method.GetServerStreaming() || method.GetClientStreaming() {
		reqArgs = append(reqArgs, servName+"_"+generator.CamelCase(origMethName)+"Stream")
	}
	if !method.GetClientStreaming() && !method.GetServerStreaming() {
		rets = append(rets, "*"+modelPkg+"."+g.typeName(method.GetOutputType()))
	}
	rets = append(rets, "error")
	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + "(" + strings.Join(rets, ", ") + ") "
}

func (g *http) generateServerMethod(servName string, method *descriptor.MethodDescriptorProto) string {
	methName := generator.CamelCase(method.GetName())
	serveType := servName + "Service"
	inType := g.typeName(method.GetInputType())
	outType := g.typeName(method.GetOutputType())

	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		GenBodyBuffer.WriteString(fmt.Sprintf("func (s *%sServer) %s(ctx %s.Context, in *%s.%s) (out *%s.%s, err error) {\n", unexport(servName), methName, contextPkg, modelPkg, inType, modelPkg, outType))
		GenBodyBuffer.WriteString(fmt.Sprintf("\treturn s.%s.%s(ctx, in)\n", serveType, methName))
		GenBodyBuffer.WriteString("}\n")
		return ""
	}

	return ""
}

type googleMethodOptionInfo struct {
	Body        string
	Method      string
	PathPattern string
	HTTPRule    *annotations.HttpRule
}

// ParseMethod TODO: NEEDS COMMENT INFO
func ParseMethod(method *descriptor.MethodDescriptorProto) (*googleMethodOptionInfo, error) {
	ext, err := proto.GetExtension(method.GetOptions(), annotations.E_Http)
	if err != nil {
		panic(err)
	}
	rule := ext.(*annotations.HttpRule)
	httpMethod := rule.GetMethod()
	pathPattern := rule.GetPattern()
	//panic(fmt.Sprintf("%s  %s", httpMethod, pathPattern))
	m := &googleMethodOptionInfo{
		Body:        rule.GetBody(),
		Method:      httpMethod,
		PathPattern: pathPattern,
		HTTPRule:    rule,
	}
	return m, nil
}

func getTagValue(key string, tags []reflect.StructTag) string {
	for _, t := range tags {
		val := t.Get(key)
		if val != "" {
			return val
		}
	}
	return ""
}

var fieldDescriptorTypeName = map[string]string{
	"TYPE_BOOL":   "bool",
	"TYPE_ENUM":   "int32",
	"TYPE_INT32":  "int32",
	"TYPE_UINT32": "uint32",

	"TYPE_INT64":  "int64",
	"TYPE_UINT64": "uint64",

	"TYPE_FLOAT":    "float64",
	"TYPE_DOUBLE":   "float64",
	"TYPE_FIXED64":  "float64",
	"TYPE_FIXED32":  "float64",
	"TYPE_SINT32":   "float64",
	"TYPE_SINT64":   "float64",
	"TYPE_SFIXED32": "float64",
	"TYPE_SFIXED64": "float64",

	"TYPE_BYTES":  "[]byte",
	"TYPE_STRING": "string",

	"TYPE_GROUP": "",
	//todo
	"TYPE_MESSAGE": "",
}
