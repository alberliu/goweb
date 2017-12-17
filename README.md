# goweb

一个基于go语言开发API的工具，这个工具受到了SpringMVC的启发，结合了go语言本身的特性，整体比较简单，接下来，看看如何使用它。

下载安装：

```
go get github.com/alberliu/goweb
```

[TOC]

### 1.核心功能

#### 请求体参数注入

```go
package main

import "github.com/alberliu/goweb"

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func handler(user User) User {
	return user
}

func main() {
	goweb.HandlePost("/test", handler)
	goweb.ListenAndServe(":8000")
}

```

请求体：

```json
{
    "id": 1,
    "name": "alber"
}
```

响应体：

```json
{
    "id": 1,
    "name": "alber"
}
```

上面的代码是一个最简的例子，HandlePost(string, interface{})会将一个handler注册到一个全局的内置的goweb实例defultGoWeb，ListenAndServe(":8000")会将defultGoWeb赋给Server的handler变量，然后启动这个Server。（是不是和内置的ServerMux有点像）

goweb会自动解析注册到它本身的handler,当请求到来时，会将请求体的json数据反序列化并注入到handler的参数，handler处理完逻辑返回时，会将handler的返回值序列化为json数据返回。goweb默认使用json的序列化和反序列化方式，当然你可以定义自己的序列化方式，这个在后面你可以看到。

例子给出的handler的参数和返回都是结构体类型，当然你也可以使用指针类型。

结构体goweb其实本质上就是一个路由，它实现了Handler接口。上面的例子都是默认的defultGoWeb，你也可以自己实例化一个goweb。

```go
func main() {
	goweb:=goweb.NewGoWeb();
	goweb.HandlePost("/test", handler)
	server := &http.Server{Addr: ":8000", Handler: goweb}
	server.ListenAndServe()
}
```



#### url参数注入

```go
package main

import "github.com/alberliu/goweb"

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func handler(id int64, name string) User {
	return User{id, name}
}

func main() {
	goweb := goweb.NewGoWeb();
	goweb.HandleGet("/test/{id}/{name}", handler)
	goweb.ListenAndServe(":8000")
}
```

执行上面的代码，然后访问url：http://localhost:8000/test/123456/alber

就可以返回下面的json数据

```json
{
    "id": 123456,
    "name": "alber"
}
```

handler可以获取到url中的参数，并且注入到handler参数中。handler的第一个参数对应url中的第一个参数，第二个参数对应url中的的第二个参数，依次类推。不过暂时还有个限制，在url中使用参数时，handler中的参数必须与url中的参数个数一致，且类型必须为string或者int64。

### 2.handler

goweb可以注册多种形式的handler，goweb会利用反射自动解析函数，支持多种类型，但是不能超出它可以解析的范围。以下是它所有能解析的类型。

```go

func handler(ctx goweb.Context) {
}

func handler(ctx goweb.Context) User {
	return User{}
}

func handler(user User) User {
	return User{}
}

func handler(ctx goweb.Context, user User) User {
	return User{}
}

func handler(name string, id int64) User {
	return User{}
}

func handler(ctx goweb.Context, name string, id int64) User {
	return User{}
}
```

Context是一个请求上下文，他只有ResponseWriter和Request两个字段，它的内部结构如下所示。你可以根据自己的需求修改源码进行扩展，例如，把它作为一个请求的会话使用。

```go
type Context struct {
	w http.ResponseWriter
	r *http.Request
}
```

### 3.用Group组织你的handler

```go
func main() {
	group1:=goweb.NewGroup("/group1")
	group1.HandleGet("/handler1",handler)
	group1.HandleGet("/handler2",handler)
	group1.HandleGet("/handler3",handler)

	group2:=goweb.NewGroup("/group2")
	group2.HandleGet("/handler1",handler)
	group2.HandleGet("/handler2",handler)
	group2.HandleGet("/handler3",handler)

	group3:=goweb.NewGroup("/group3")
	group3.HandleGet("/handler1",handler)
	group3.HandleGet("/handler2",handler)
	group3.HandleGet("/handler3",handler)

	goweb.HandleGroup(group1)
	goweb.HandleGroup(group2)
	goweb.HandleGroup(group3)
	goweb.ListenAndServe(":8000")
}
```

group可以帮助你分层次的组织你的handler,使你的路由结构更清晰。

### 4.定义自己序列化和反序列话方式

```go
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func jsonMarshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func main() {
	goweb:=goweb.NewGoWeb();
	goweb.Unmarshal=jsonUnmarshal
	goweb.Marshal=jsonMarshal
	
	goweb.ListenAndServe(":8000")
	
}
```

goweb默认采用json(使用的是开源的jsoniter)序列化和反序列化数据，goweb的Marshal、Unmarshal变量本身是一个函数.如果你想定义自己的序列化方式，只需要覆盖掉它就行，就像上面那样。

### 5.拦截器

```go

func interceptor1(http.ResponseWriter, *http.Request) bool {
	return true
}
func interceptor2(http.ResponseWriter, *http.Request) bool {
	return true
}
func interceptor3(http.ResponseWriter, *http.Request) bool {
	return true
}

func main() {
	goweb := goweb.NewGoWeb();
	goweb.AddInterceptor(interceptor1)
	goweb.AddInterceptor(interceptor2)
	goweb.AddInterceptor(interceptor3)
	goweb.ListenAndServe(":8000")
}

```

goweb在执行handler之前，会执行一个或者多个interceptor，并且会根据AddInterceptor的先后顺序执行，当interceptor返回true时，会接着往下执行，返回false时，会终止执行。

### 6.过滤器

```
func filter(w http.ResponseWriter, r *http.Request, f func(http.ResponseWriter, *http.Request)) {
	f(w, r)
}

func main() {
	goweb := goweb.NewGoWeb();
	goweb.Filter = filter
	goweb.ListenAndServe(":8000")
}

```

你可以给goweb添加一个过略器，在过滤器中，如果你想执行完自己的逻辑之后，执行handler，只需要调用f(w, r)。

### 7.自定义错误处理

```go

func handler400(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	w.Write([]byte("bad request"))
}
func handler404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("url not found"))
}
func handler405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	w.Write([]byte("method not found"))
}

func main() {
	goWeb := goweb.NewGoWeb()
	goWeb.Handler400 = handler400
	goWeb.Handler404 = handler404
	goWeb.Handler405 = handler405

	goweb.ListenAndServe(":8000")
}

```

当请求执行失败时，goweb中给出了一些默认的错误处理方式，就像上面那样。当然，你也可以定义一些自己错误处理方式。

### 写在后面

如果你有什么好的建议，可以发我邮箱，一起交流。

alber_liu@qq.com



