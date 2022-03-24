package pkg

import (
	"fmt"
	"strings"
)

// GoImport Use golang.org/x/tools/imports auto import pkg
func GoImport(file string, bytes []byte) (res []byte, err error) {
	options := &Options{
		AllErrors:  false,
		TabWidth:   8,
		TabIndent:  true,
		Comments:   true,
		Fragment:   true,
		FormatOnly: false,
	}
	if res, err = Process(file, bytes, options); err != nil {
		fmt.Printf("GoImport(%s) error(%v)", file, err)
		res = bytes
		return
	}
	return
}

// ConvertMethod checkout the file belongs to dao or not
func ConvertMethod(path string) (method string) {
	switch {
	case strings.Contains(path, "/dao"):
		method = "d"
	case strings.Contains(path, "/service"):
		method = "s"
	case strings.Contains(path, "/manager"):
		method = "m"
	default:
		method = ""
	}
	return
}

// ConvertHump convert words to hump style
func ConvertHump(words string) string {
	return strings.ToUpper(words[0:1]) + words[1:]
}

// RmDup remove duplicate element of []string
func RmDup(list []string) []string {
	if len(list) == 0 {
		return []string{}
	}
	m := make(map[string]bool)
	result := make([]string, 0)
	for _, v := range list {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}
