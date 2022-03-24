package goparser

import (
	"fmt"
	"go/importer"
	"testing"

	"git.inke.cn/BackendPlatform/daenerys-tool/daenerys/internal/models"
)

func TestParser_Parse(t *testing.T) {
	p := Parser{Importer: importer.Default()}
	result, err := p.Parse("/Users/melody/go/src/git.inke.cn/FrtMeLody/http-example/server/http/http.go", []models.Path{"/Users/melody/go/src/git.inke.cn/FrtMeLody/http-example/server/http/handler.go"})
	_ = err
	for _, name := range result.Funcs {
		fmt.Printf("%+v\n", name.Name)
	}
	// for _, h := range result.Header.Comments {
	// 	fmt.Println(h)
	// }
	// fmt.Println(result.Header.Package)
	// for _, imp := range result.Header.Imports {
	// 	fmt.Println(imp)
	// }
	// fmt.Println(string(result.Header.Code))
	// // for _, h := range result.Header.Package{
	// fmt.Println(h)
	// }
}
