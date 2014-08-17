package main

import (
	"fmt"
	"time"

	"github.com/hprose/hprose-go/hprose"
)

var (
	gClient    hprose.Client
	gS         = new(gStub)
	gSessionId = ""
	gIndex     = 0
	gUrl       = ""
	active     = false
)

type gStub struct {
	Auth        func(string) (string, int, error)
	SwitchReady func(string) error
	SendChat    func(string, string) error
	Operate     func(string, string) error
	Quit        func(string) error
	Ping        func(string) error
	GetData     func(int, string) ([]string, int, error)
}

func joinGameServer(host string, token string) {
	url := fmt.Sprintf("http://%s/", host)
	fmt.Println("connected to rpc server:", url)
	gClient = hprose.NewHttpClient(url)
	gClient.SetUri(url)
	gClient.UseService(&gS)

	var err error
	gSessionId, gIndex, err = gS.Auth(token)
	if err != nil {
		fmt.Printf("get error when validate token: %v\n", err)
		return
	}
	active = true
	go ping()
	go getData()

	for {
		time.Sleep(5 * time.Second)
		if err := gS.SendChat(fmt.Sprintf("hello all, now time is %v", time.Now().Unix()), gSessionId); err != nil {
			fmt.Printf("can not send chat: %v", err)
			active = false
			return
		}
	}
}

func ping() {
	fmt.Printf("start ping with interval %d seconds\n", 3)
	for {
		time.Sleep(3 * time.Second)
		if !active {
			fmt.Printf("stop ping\n")
			return
		}
		if err := gS.Ping(gSessionId); err != nil {
			active = false
			fmt.Printf("ping error: %v\n", err)
			return
		}
	}
}

func getData() {
	for {
		if !active {
			fmt.Printf("stop getting data\n")
			return
		}
		vals, i, err := gS.GetData(gIndex, gSessionId)
		if err != nil {
			fmt.Printf("can not get data: %v\n", err)
			active = false
			return
		}
		gIndex = i
		for _, v := range vals {
			fmt.Printf("get data %s\n", v)
		}
	}
}
