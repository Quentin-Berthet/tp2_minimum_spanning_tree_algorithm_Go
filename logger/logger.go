package logger

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func GetFunctionName(i interface{}) string {
	namespace := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")
	return namespace[len(namespace) - 1]
}

func Log(function interface{}, content string) {
	fmt.Printf("[%s] %s\n", GetFunctionName(function), content)
}
