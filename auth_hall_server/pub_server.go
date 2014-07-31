package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

var httpPubServer = hprose.NewHttpService()

var pubServerEnable = true

type (
	pubStub struct{}
	pubSe   struct{}
)

func (pubSe) OnBeforeInvoke(fName string, params []reflect.Value, isSimple bool, ctx interface{}) {
	if !pubServerEnable {
		panic("we are closing the server, not accept request at the moment")
	}

	log.Info("create session for ip: %s", utils.GetIp(ctx))
	session.CreateSession(ctx)

	str := "invoker from " + utils.GetIp(ctx) + "\n"
	str += "invokes the function " + fName + "\n"
	str += "parameters: "
	for _, v := range params {
		str += fmt.Sprintf("%v  ", v.Interface())
	}
	str += "\n"
	log.Info(str)
}

func (pubSe) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {}

func (pubSe) OnSendError(error, interface{}) {}

func initPubServer() {
	httpPubServer.AddMethods(pubStub{})
	httpPubServer.ServiceEvent = pubSe{}
	httpPubServer.CrossDomainEnabled = true
	// httpServer.AddAccessControlAllowOrigin()
	go servePubHttp()
}

func servePubHttp() {
	if err := http.ListenAndServe(":"+pubRpcPort, httpPubServer); err != nil {
		panic(err)
	}
}