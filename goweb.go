package goweb

import (
	"net/http"
	"reflect"
	"errors"
	"sort"
	"strconv"
	"strings"
	"log"
)

const DEFULT_NUM_HANDLER = 10

type method uint8

const (
	GET    method = iota
	POST
	PUT
	DELETE
)

var defultGoWeb = NewGoWeb()

type goWeb struct {
	router       *router                                                                            //路由
	interceptors []func(http.ResponseWriter, *http.Request) bool                                    //拦截器
	Filter       func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request)) //过略器
	Unmarshal    func([]byte, interface{}) error                                                    //解码
	Marshal      func(interface{}) ([]byte, error)                                                  //编码
	Handler400   func(http.ResponseWriter, *http.Request)                                           //处理400错误
	Handler404   func(http.ResponseWriter, *http.Request)                                           //处理404异常
	Handler405   func(http.ResponseWriter, *http.Request)                                           //处理405异常
}

func NewGoWeb() *goWeb {
	gw := goWeb{}
	gw.router = newRouter()
	gw.interceptors =make([]func(http.ResponseWriter, *http.Request) bool,0,5)
	gw.Unmarshal = jsonUnmarshal
	gw.Marshal = jsonMarshal
	gw.Handler400 = handler400
	gw.Handler404 = handler404
	gw.Handler405 = handler405
	return &gw
}

//添加拦截器
func (gw *goWeb)AddInterceptor(interceptor func(http.ResponseWriter, *http.Request) bool){
	gw.interceptors=append(gw.interceptors,interceptor)
}

//注册handler组
func (gw *goWeb) HandleGroup(group *group) {
	count := strings.Count(group.gUrl, "/")
	for _, v := range group.controllers {
		v.url = group.gUrl + v.url
		if v.indexUrlparam != nil {
			for i, _ := range v.indexUrlparam {
				v.indexUrlparam[i] = v.indexUrlparam[i] + count
			}
		}
		gw.router.addController(v)
	}
	sort.Sort(gw.router)
}

//注册handler
func (gw *goWeb) HandleFunc(method method, url string, handler interface{}) {
	handlerType, indexUrlParam := checkHandler(url, handler)
	if method == GET && (handlerType == BODY_INJECT || handlerType == BODY_INJECT_CTX) {
		panic("handler error")
	}
	controller := controller{
		method:        method,
		url:           url,
		handler:       handler,
		handlerType:   handlerType,
		indexUrlparam: indexUrlParam,
	}
	gw.router.addController(controller)
	sort.Sort(gw.router)
}

//注册处理GET请求的handler
func (gw *goWeb) HandleGet(url string, handler interface{}) {
	gw.HandleFunc(GET, url, handler)
}

//注册处理POST请求的handler
func (gw *goWeb) HandlePost(url string, handler interface{}) {
	gw.HandleFunc(POST, url, handler)
}

//注册处理PUT请求的handler
func (gw *goWeb) HandlePut(url string, handler interface{}) {
	gw.HandleFunc(PUT, url, handler)
}

//注册处理DELETE请求的handler
func (gw *goWeb) HandleDelete(url string, handler interface{}) {
	gw.HandleFunc(DELETE, url, handler)
}

//启动一个服务器
func (gw *goWeb) ListenAndServe(addr string) error {
	server := &http.Server{Addr: addr, Handler: gw}
	return server.ListenAndServe()
}


func HandleGroup(group *group) {
	defultGoWeb.HandleGroup(group)
}

func HandleFunc(url string, method method, handler interface{}) {
	defultGoWeb.HandleFunc(method, url, handler)
}

func HandleGet(url string, handler interface{}) {
	defultGoWeb.HandleGet(url, handler)
}

func HandlePost(url string, handler interface{}) {
	defultGoWeb.HandlePost(url, handler)
}

func HandlePut(url string, handler interface{}) {
	defultGoWeb.HandlePut(url, handler)
}

func HandleDelete(url string, handler interface{}) {
	defultGoWeb.HandleDelete(url, handler)
}

func ListenAndServe(addr string) error {
	return defultGoWeb.ListenAndServe(addr)
}

//实现Handler接口
func (gw *goWeb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//执行拦截器
	if gw.interceptors != nil {
		for _, v := range gw.interceptors {
			if v(w, r) {
				continue
			}
			return
		}
	}

	//执行过滤器
	h := gw.handleHttp
	if gw.Filter != nil {
		gw.Filter(w, r, h)
		return
	}

	h(w, r)
}

//执行注册的handler
func (gw *goWeb) handleHttp(w http.ResponseWriter, r *http.Request) {
	method, err := getMethod(r.Method)
	if err != nil {
		gw.Handler405(w, r)
		return
	}

	url := r.RequestURI
	controller := gw.router.Search(url)
	if (controller == nil) {
		gw.Handler404(w, r)
		return
	}
	if method != controller.method {
		gw.Handler405(w, r)
		return
	}

	switch controller.handlerType {
	case SINPLE:
		gw.doSinple(controller, w, r)
	case SINPLE_RETURN:
		gw.doSinpleReturn(controller, w, r)
	case BODY_INJECT:
		gw.doBodyInject(controller, w, r)
	case BODY_INJECT_CTX:
		gw.doBodyInjectCtx(controller, w, r)
	case URL_INJECT:
		gw.doUrlInject(controller, w, r)
	case URL_INJECT_CTX:
		gw.doUrlInjectCtx(controller, w, r)
	}
}

//处理简单类型
func (gw *goWeb) doSinple(controller *controller, w http.ResponseWriter, r *http.Request) {
	handlerValue := reflect.ValueOf(controller.handler)
	params := make([]reflect.Value, 1)
	ctx := Context{w, r}
	params[0] = reflect.ValueOf(ctx)
	handlerValue.Call(params)
}

//处理带返回值的简单类型
func (gw *goWeb) doSinpleReturn(controller *controller, w http.ResponseWriter, r *http.Request) {
	handlerValue := reflect.ValueOf(controller.handler)
	params := make([]reflect.Value, 1)
	ctx := Context{w, r}
	params[0] = reflect.ValueOf(ctx)
	returnSlice := handlerValue.Call(params)

	//获取处理后的返回值
	return1 := returnSlice[0]
	jsonResponse, err := gw.Marshal(return1.Interface())
	log.Println("marshal error:",err)
	w.Write(jsonResponse)
}

//处理请求体注入类型
func (gw *goWeb) doBodyInject(controller *controller, w http.ResponseWriter, r *http.Request) {
	handler := controller.handler

	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	//获取handler的第一个参数,并且实例化一个结构体
	paramType := handlerType.In(0)

	var paramValue reflect.Value
	isPtr := paramType.Kind() == reflect.Ptr
	if isPtr {
		paramValue = reflect.New(paramType.Elem())
	} else {
		paramValue = reflect.New(paramType)
	}

	pParam := paramValue.Interface()

	//获取请求体的json，并且解析到handler的参数中
	bytes := make([]byte, r.ContentLength)
	r.Body.Read(bytes)

	err := gw.Unmarshal(bytes, pParam)
	if err != nil {
		gw.Handler400(w, r)
		return
	}

	//构造第一个参数
	params := make([]reflect.Value, 1)
	if isPtr {
		params[0] = paramValue
	} else {
		params[0] = paramValue.Elem()
	}

	//调用handler处理请求
	returnSlice := handlerValue.Call(params)

	//获取处理后的返回值
	return1 := returnSlice[0]
	jsonResponse, err := gw.Marshal(return1.Interface())
	log.Println("marshal error:",err)
	w.Write(jsonResponse)
}

//处理携带请求上下文的请求体注入类型
func (gw *goWeb) doBodyInjectCtx(controller *controller, w http.ResponseWriter, r *http.Request) {
	handler := controller.handler

	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	//获取handler的第一个参数,并且实例化一个结构体
	paramType := handlerType.In(1)

	var paramValue reflect.Value
	isPtr := paramType.Kind() == reflect.Ptr
	if isPtr {
		paramValue = reflect.New(paramType.Elem())
	} else {
		paramValue = reflect.New(paramType)
	}

	pParam := paramValue.Interface()

	//获取请求体的json，并且解析到handler的参数中
	bytes := make([]byte, r.ContentLength)
	r.Body.Read(bytes)

	error := gw.Unmarshal(bytes, pParam)
	if error != nil {
		gw.Handler404(w, r)
		return
	}

	//构造参数
	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(Context{w, r})
	if isPtr {
		params[1] = paramValue
	} else {
		params[1] = paramValue.Elem()
	}

	//调用handler处理请求
	returnSlice := handlerValue.Call(params)

	//获取处理后的返回值
	return1 := returnSlice[0]
	jsonResponse, err := gw.Marshal(return1.Interface())
	log.Println("marshal error:",err)
	w.Write(jsonResponse)
}

//处理url参数注入类型
func (gw *goWeb) doUrlInject(controller *controller, w http.ResponseWriter, r *http.Request) {
	rurl := r.RequestURI

	handler := controller.handler
	handlerType := reflect.TypeOf(handler);
	num := handlerType.NumIn()

	params := make([]reflect.Value, num)

	splists := strings.Split(rurl, "/")

	for i := 0; i < num; i++ {
		str := splists[controller.indexUrlparam[i]+1] //+1舍弃第一个空字符
		if handlerType.In(i).Kind() == reflect.String {
			params[i] = reflect.ValueOf(str)
			continue
		}
		strInt, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			gw.Handler400(w, r)
			return
		}
		params[i] = reflect.ValueOf(strInt)
	}

	handlerValue := reflect.ValueOf(handler)
	returnSlice := handlerValue.Call(params)

	//获取处理后的返回值
	return1 := returnSlice[0]
	jsonResponse, err := json.Marshal(return1.Interface())
	log.Println("marshal error:",err)
	w.Write(jsonResponse)
}

//处理携带请求上下文的url参数注入类型
func (gw *goWeb) doUrlInjectCtx(controller *controller, w http.ResponseWriter, r *http.Request) {
	rurl := r.RequestURI

	handler := controller.handler
	handlerType := reflect.TypeOf(handler);
	num := handlerType.NumIn()

	params := make([]reflect.Value, num)
	params[0] = reflect.ValueOf(Context{w, r})
	splists := strings.Split(rurl, "/")
	for i := 1; i < num; i++ {
		str := splists[controller.indexUrlparam[i-1]+1]
		if handlerType.In(i).Kind() == reflect.String {
			params[i] = reflect.ValueOf(str)
			continue
		}
		strInt, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			gw.Handler400(w, r)
			return
		}
		params[i] = reflect.ValueOf(strInt)
	}

	handlerValue := reflect.ValueOf(handler)
	returnSlice := handlerValue.Call(params)

	//获取处理后的返回值
	return1 := returnSlice[0]
	jsonResponse, err := gw.Marshal(return1.Interface())
	log.Println("marshal error:",err)
	w.Write(jsonResponse)
}

//获取请求的方法
func getMethod(method string) (method, error) {
	switch method {
	case "GET":
		return GET, nil
	case "POST":
		return POST, nil
	case "PUT":
		return PUT, nil
	case "DELETE":
		return DELETE, nil
	default:
		return 0, errors.New("method not surport")
	}
}
