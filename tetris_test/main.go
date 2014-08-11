package main

import (
	"fmt"
	"reflect"
	"strconv"
)

var t reflect.Type
var v = []reflect.Value{reflect.ValueOf(fm{})}

func init() {
	t = reflect.TypeOf(fm{})
	initRpcClient()
}

func main() { handler() }

func handler() {
	for {
		fmt.Println(option())
		var op string
		if _, err := fmt.Scanln(&op); err != nil {
			fmt.Println("can not scan op: ", err)
			continue
		}
		opCode, err := strconv.Atoi(op)
		if err != nil {
			fmt.Println("can not convert string to int: ", op)
			continue
		}
		t.Method(opCode).Func.Call(v)
	}
}

func option() string {
	str := "choose an option:\n"
	for i := 0; i < t.NumMethod(); i++ {
		str += fmt.Sprintf("%d. %s\n", i, t.Method(i).Name)
	}
	return str
}

type fm struct{}

func (fm) Register()           { register() }
func (fm) Login()              { login() }
func (fm) Logout()             { logout() }
func (fm) GetUserInfo()        { getUserInfo() }
func (fm) Create()             { create() }
func (fm) CreateSession()      { createSession() }
func (fm) GetNormalHall()      { getNormal() }
func (fm) Join()               { join() }
func (fm) UpdateUserAvatar()   { updateUserAvatar() }
func (fm) UpdateUserPassword() { updateUserPassword() }
