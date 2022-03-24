package format

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func FmtUnmarshalStruct(str string, target interface{}) error {
	if len(str) == 0 {
		return fmt.Errorf("empty")
	}

	if strings.HasSuffix(str, "&") {
		str = str[1:]
	}

	if strings.HasPrefix(str, "{") {
		str = str[1:]
	}

	if strings.HasSuffix(str, "}") {
		str = str[:len(str)-1]
	}

	seg := strings.Split(str, " ")

	index := 0
	getKv := func(more int) (string, error) {
		parts := strings.SplitN(seg[index], ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid seg: %s", seg[index])
		}
		value := parts[1]
		index++

		if more > 0 {
			other := make([]string, 0, more+1)
			other = append(other, value)
			other = append(other, seg[index:index+more]...)
			index += more
			value = strings.Join(other, " ")
		}

		return value, nil
	}

	v := reflect.ValueOf(target).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		val, err := StringToKind(field.Type(), getKv)
		if err != nil {
			return fmt.Errorf("unsupport field kind %s err: %v", field.Kind(), err)
		}
		if val != nil {
			field.Set(val.Convert(field.Type()))
		}
	}

	return nil
}

var (
	timeKind = reflect.TypeOf(time.Time{}).Kind()

	KindParseMap = map[reflect.Kind]func(string) (interface{}, error){
		timeKind: func(s string) (interface{}, error) {
			return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", s)
		},
		reflect.String: func(s string) (interface{}, error) {
			return s, nil
		},
		reflect.Uint: func(s string) (interface{}, error) {
			v, err := strconv.ParseUint(s, 10, 8)
			return uint(v), err
		},
		reflect.Uint8: func(s string) (interface{}, error) {
			v, err := strconv.ParseUint(s, 10, 8)
			return uint8(v), err
		},
		reflect.Uint16: func(s string) (interface{}, error) {
			v, err := strconv.ParseUint(s, 10, 16)
			return uint16(v), err
		},
		reflect.Uint32: func(s string) (interface{}, error) {
			v, err := strconv.ParseUint(s, 10, 32)
			return uint32(v), err
		},
		reflect.Uint64: func(s string) (interface{}, error) {
			return strconv.ParseUint(s, 10, 64)
		},
		reflect.Int: func(s string) (interface{}, error) {
			v, err := strconv.ParseInt(s, 10, 64)
			return int(v), err
		},
		reflect.Int8: func(s string) (interface{}, error) {
			v, err := strconv.ParseInt(s, 10, 8)
			return int8(v), err
		},
		reflect.Int16: func(s string) (interface{}, error) {
			v, err := strconv.ParseInt(s, 10, 16)
			return int16(v), err
		},
		reflect.Int32: func(s string) (interface{}, error) {
			v, err := strconv.ParseInt(s, 10, 32)
			return int32(v), err
		},
		reflect.Int64: func(s string) (interface{}, error) {
			return strconv.ParseInt(s, 10, 64)
		},
	}

	KindSegLenMap = map[reflect.Kind]int{
		timeKind: 3,
	}
)

func StringToKind(ty reflect.Type, f func(int) (string, error)) (*reflect.Value, error) {
	kind := ty.Kind()

	str, err := f(KindSegLenMap[kind])
	if err != nil {
		return nil, err
	}
	if len(str) == 0 {
		return nil, nil
	}

	parser, ok := KindParseMap[kind]
	if !ok {
		return nil, fmt.Errorf("unsupport kind: %s", kind.String())
	}

	val, err := parser(str)
	if err != nil {
		return nil, fmt.Errorf("kind %s parse err: %v", kind.String(), err)
	}

	v := reflect.ValueOf(val)
	return &v, nil
}
