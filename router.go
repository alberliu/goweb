package goweb

import (
	"sort"
	"reflect"
)

//handler的类型,项目启动时候解析，避免请求时解析
type handlerType uint8

const (
	SINPLE          handlerType = iota //简单模型
	SINPLE_RETURN                      //简单模型，有返回值
	BODY_INJECT                        //请求体注入类型
	BODY_INJECT_CTX                    //请求体注入类型，携带请求上下文
	URL_INJECT                         //URL注入类型
	URL_INJECT_CTX                     //URL注入类型，携带上下文
)

type router struct {
	controllers []controller
}

type controller struct {
	method        method      //请求的方法
	url           string      //请求url
	handler       interface{} //处理的函数
	handlerType   handlerType //handler处理类型
	indexUrlparam []int       //url中参数的位置
}

func newRouter() *router {
	router := router{}
	router.controllers = make([]controller, 0, DEFULT_NUM_HANDLER)
	return &router
}

func (rt *router) addController(controller controller) {
	rt.controllers = append(rt.controllers, controller)
}

//实现排序接口
func (rt *router) Len() int {
	return len(rt.controllers)
}

func (rt *router) Less(i, j int) bool {
	if compare(rt.controllers[i].url, rt.controllers[j].url) == -1 {
		return true
	}
	return false
}

func (rt *router) Swap(i, j int) {
	rt.controllers[i], rt.controllers[j] = rt.controllers[j], rt.controllers[i]
}

//二分查找
func (rt *router) Search(url string) (*controller) {
	f := func(i int) bool {
		if compare(rt.controllers[i].url, url) == -1 {
			return false
		}
		return true
	}
	len := len(rt.controllers)
	index := sort.Search(len, f)
	if index != len && compare(rt.controllers[index].url, url) == 0 {
		return &rt.controllers[index]
	}
	return nil
}

//handler校验
func checkHandler(url string, handler interface{}) (handlerType, []int) {
	var (
		result     = false
		iUrlParams []int
	)

	num := numUrlParam(url)
	if num == -1 {
		panic("illegal url")
	}

	if num == 0 {
		if isSinple(handler) {
			return SINPLE, nil
		}
		if isSinpleReturn(handler) {
			return SINPLE_RETURN, nil
		}

		if isBodyInject(handler) {
			return BODY_INJECT, nil
		}

		if isBodyInjectCtx(handler) {
			return BODY_INJECT_CTX, nil
		}
		panic("handler error")
	}

	result, iUrlParams = isUrlInject(url, handler)
	if result {
		return URL_INJECT, iUrlParams
	}

	result, iUrlParams = isUrlInjectCtx(url, handler)
	if result {
		return URL_INJECT_CTX, iUrlParams
	}
	panic("handler error")
}

//是否是简单类型
func isSinple(handler interface{}) bool {
	var f func(cxt Context)
	if reflect.TypeOf(f) == reflect.TypeOf(handler) {
		return true
	}
	return false
}

//是否是有返回值的简单类型
func isSinpleReturn(handler interface{}) bool {
	handlerType := reflect.TypeOf(handler)
	if handlerType.NumIn() == 1 && handlerType.In(0) == typeContext &&
		handlerType.NumOut() == 1 {
		return true;
	}
	return false
}

//是否是请求体注入类型
func isBodyInject(handler interface{}) (bool) {
	handlerType := reflect.TypeOf(handler)
	if handlerType.NumIn() == 1 && handlerType.NumOut() == 1 {
		return true
	}
	return false;
}

//是否是携带请求上下文的请求体注入类型
func isBodyInjectCtx(handler interface{}) (bool) {
	handlerType := reflect.TypeOf(handler)
	if handlerType.NumIn() == 2 && handlerType.In(0) == typeContext &&
		handlerType.NumOut() == 1 {
		return true
	}
	return true;
}

//是否是url参数注入类型
func isUrlInject(url string, handler interface{}) (bool, []int) {
	handlerType := reflect.TypeOf(handler)

	num := numUrlParam(url)
	if num != handlerType.NumIn() {
		return false, nil
	}

	for i := 0; i < handlerType.NumOut(); i++ {
		paramKind := handlerType.In(i).Kind()
		if paramKind != reflect.Int64 && paramKind != reflect.String {
			return false, nil
		}
	}

	if handlerType.NumOut() != 1 {
		return false, nil
	}

	iUrlParam := make([]int, num)
	indexUrlParam(url, iUrlParam)
	return true, iUrlParam
}

//是否是携带请求上下文的url参数注入类型
func isUrlInjectCtx(url string, handler interface{}) (bool, []int) {
	handlerType := reflect.TypeOf(handler)

	if handlerType.In(0) != typeContext {
		return false, nil
	}

	num := numUrlParam(url)
	if num != handlerType.NumIn()-1 {
		return false, nil
	}

	for i := 1; i < handlerType.NumOut(); i++ {
		paramKind := handlerType.In(i).Kind()
		if paramKind != reflect.Int64 && paramKind != reflect.String {
			return false, nil
		}
	}

	if handlerType.NumOut() != 1 {
		return false, nil
	}

	iUrlParam := make([]int, num)
	indexUrlParam(url, iUrlParam)
	return true, iUrlParam
}

//url中参数个数
func numUrlParam(str string) (num int) {
	if str == "" {
		return -1
	}
	num = 0;
	flag := 0
	for i := 0; i < len(str); i++ {
		if str[i] == '{' {
			if str[i-1] != '/' {
				return -1
			}
			if flag != 0 {
				return -1
			}
			flag = 1
		}
		if str[i] == '}' {
			if i+1 != len(str) && str[i+1] != '/' {
				return -1
			}
			if flag != 1 {
				return -1
			}
			flag = 0
			num++
		}
	}
	return num
}

//url中参数位置
func indexUrlParam(url string, indexUrlParam []int) {
	iSlash := 0
	iBraces := 0
	for i := 0; i < len(url); i++ {
		if url[i] == '/' {
			iSlash++
		}
		if url[i] == '{' {
			indexUrlParam[iBraces] = iSlash - 1
			iBraces++
		}
	}
}

//url匹配算法，支持restful接口
func compare(s1, s2 string) int {
	if s1 == "" || s2 == "" {
		panic("str is null")
	}
	lens1 := len(s1)
	lens2 := len(s2)

	for i, j := 0, 0; ; {
		if i == lens1 && j == lens2 {
			return 0
		}
		if i == lens1 {
			return -1
		}
		if j == lens2 {
			return 1
		}

		if (s1[i] == '{') {
			i = index(s1[i:], '}') + i + 1
			j = index(s2[j:], '/') + j
			continue
		}
		if (s2[j] == '{') {
			i = index(s1[i:], '/') + i
			j = index(s2[j:], '}') + j + 1
			continue
		}

		if s1[i] == s2[j] {
			i++
			j++
			continue
		}
		if s1[i] > s2[j] {
			return 1
		}
		if s1[i] < s2[j] {
			return -1
		}

	}
}

//返回字符c在字符串str中第一次出现的位置，如果没有找到，就返回字符串的长度
func index(str string, c byte) int {
	lstr := len(str)
	for i := 0; i < lstr; i++ {
		if str[i] == c {
			return i
		}
	}
	return lstr
}
