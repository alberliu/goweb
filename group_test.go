package goweb

import (
	"testing"
	"fmt"
)

func TestGroup_HandleGet(t *testing.T) {
	group := NewGroup("/hello")
	group.HandleGet("/", handler)
	fmt.Println(group)
}

func TestGroup_HandlePost(t *testing.T) {
	group := NewGroup("/hello")
	group.HandlePost("/", handler)
	fmt.Println(group)
}

func TestGroup_HandlePut(t *testing.T) {
	group := NewGroup("/hello")
	group.HandlePut("/", handler)
	fmt.Println(group)
}

func TestGroup_HandleDelete(t *testing.T) {
	group := NewGroup("/hello")
	group.HandleDelete("/", handler)
	fmt.Println(group)
}

