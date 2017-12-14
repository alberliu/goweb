package goweb

type group struct {
	gUrl        string
	controllers []controller
}

func NewGroup(gurl string) *group {
	group := group{}
	group.gUrl = gurl
	group.controllers = make([]controller, 0, DEFULT_NUM_HANDLER)
	return &group
}

func (group *group) addController(controller controller) {
	group.controllers = append(group.controllers, controller)
}

func (group *group) HandleFunc(method method, url string, handler interface{}) {
	handlerType, indexUrlParam := checkHandler(url, handler)
	if method == GET && handlerType == BODY_INJECT {
		panic("handler error")
	}

	controller := controller{
		method:        GET,
		url:           url,
		handler:       handler,
		handlerType:   handlerType,
		indexUrlparam: indexUrlParam,
	}
	group.addController(controller)
	return
}

//请求方法和url
func (group *group) HandleGet(url string, hander interface{}) {
	group.HandleFunc(GET, url, hander)
}
func (group *group) HandlePost(url string, hander interface{}) {
	group.HandleFunc(POST, url, hander)
}
func (group group) HandlePut(url string, hander interface{}) {
	group.HandleFunc(PUT, url, hander)
}
func (group *group) HandleDelete(url string, hander interface{}) {
	group.HandleFunc(DELETE, url, hander)
}
