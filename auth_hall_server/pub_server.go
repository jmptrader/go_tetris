package main

import (
	"net/http"
	"reflect"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

var (
	pubServerFrequencyLimit = utils.NewFrequency(120, 5) // request to rpc server can not exceeds 120 times per minute
	httpPubServer           = hprose.NewHttpService()
	errCreateSessionFirst   = "请先创建session"
	errClosingServer        = "we are closing the server, not accept request at the moment"
	pubServerEnable         = true
)

type (
	pubStub struct{}
	pubSe   struct{}
)

func (pubSe) OnBeforeInvoke(fName string, params []reflect.Value, isSimple bool, ctx interface{}) {
	if !pubServerEnable {
		panic(errClosingServer)
	}

	if !*debug {
		if err := pubServerFrequencyLimit.Incr(utils.GetIp(ctx)); err != nil {
			panic(err)
		}
	}

	log.Info(utils.HproseLog(fName, params, ctx))

	if l := len(params); l > 0 {
		if !notNeedSessFunc[fName] && !session.IsSessIdExist(params[l-1].String()) {
			panic(errCreateSessionFirst)
		}
	}
}

func (pubSe) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {}

func (pubSe) OnSendError(error, interface{}) {}

func initPubServer() {
	httpPubServer.DebugEnabled = *debug
	httpPubServer.AddMethods(pubStub{})
	httpPubServer.ServiceEvent = pubSe{}
	httpPubServer.CrossDomainEnabled = true
	httpPubServer.SetCrossDomainXmlFile(crossDomainFile)
	go servePubHttp()
}

func servePubHttp() {
	if err := http.ListenAndServe(":"+pubRpcPort, httpPubServer); err != nil {
		panic(err)
	}
}
