/*
	rpc server for users
*/
package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

var pubHttpServer = hprose.NewHttpService()

const sessKeyPing = "ping"

type pubStub struct{}

type pubSe struct{}

var checkSessionId = func(params []reflect.Value) {
	if l := len(params); l > 0 {
		sessId := params[l-1].String()
		if !session.IsSessIdExist(sessId) {
			panic("先调用ValidateToken 创建游戏服务器上的sessionId才能发送指令")
		}
		session.SetSession(sessKeyPing, 0, sessId)
	}
}

var panicOfServerStatus = func() {
	if serverStatus != statusActive {
		panic("游戏服务器暂时不服务")
	}
}

func (pubSe) OnBeforeInvoke(funcName string, params []reflect.Value, isSimple bool, ctx interface{}) {
	log.Info(utils.HproseLog(funcName, params, ctx))

	switch funcName {
	case "ValidateToken":
		panicOfServerStatus()
	case "SwitchReady":
		checkSessionId(params)
		panicOfServerStatus()
	case "SendChat":
		checkSessionId(params)
		panicOfServerStatus()
	case "Operate":
		checkSessionId(params)
	case "Quit":
		checkSessionId(params)
	case "Ping":
		checkSessionId(params)
		panicOfServerStatus()
	case "GetData":
		checkSessionId(params)
		panicOfServerStatus()
	default:
		log.Info("what is missing? -> %s", funcName)
	}
}

func (pubSe) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {
}

func (pubSe) OnSendError(error, interface{}) {
}

func initPubRpcServer() {
	pubHttpServer.AddMethods(pubStub{})
	pubHttpServer.DebugEnabled = true
	pubHttpServer.ServiceEvent = pubSe{}
	go servePubHttp()
}

func servePubHttp() {
	if err := http.ListenAndServe(fmt.Sprintf(":%s", gamePubServerRpcPort), pubHttpServer); err != nil {
		panic(err)
	}
}
