package goweb

import (
	"testing"
	"fmt"
	"sort"
)

func TestRouter(t *testing.T) {
	router := newRouter()
	router.addController(controller{GET, "3", nil, SINPLE, nil})
	router.addController(controller{GET, "1", nil, SINPLE, nil})
	router.addController(controller{GET, "5", nil, SINPLE, nil})
	router.addController(controller{GET, "0", nil, SINPLE, nil})
	router.addController(controller{GET, "0", nil, SINPLE, nil})
	router.addController(controller{GET, "5", nil, SINPLE, nil})

	sort.Sort(router)

	fmt.Println(router)

	fmt.Println(router.Search("2"))
}

func TestCompare(t *testing.T) {
	fmt.Println(compare("1", "1"))
}

func TestNumUrlParam(t *testing.T) {
	fmt.Println(numUrlParam("/{}/{}/{}l/"))
}

func TestIndexUrlParam(t *testing.T) {
	slice := make([]int, 5)
	indexUrlParam("/{}/{jk}/kl/{}", slice)
	fmt.Println(slice)
}

type S1 struct {
}

func h0(Context){

}
func h1(ctx Context)S1{
	return S1{}
}

func h2(s *S1)S1{
	return S1{}
}

func h3(ctx Context,s *S1)S1{
	return S1{}
}

func h4(name string,id int64)S1{
	return S1{}
}

func h5(ctx Context,name string,id int64)S1{
	return S1{}
}

func TestCheckHandler(t *testing.T) {
	ty, index := checkHandler("/{alber}/{123456}", h5)
	fmt.Println(ty, index)
}
