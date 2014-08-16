package main

import "github.com/hprose/hprose-go/hprose"

var (
	gClient    hprose.Client
	gS         = new(gStub)
	gSessionId = ""
	gUrl       = ""
)

type gStub struct {
	ValidateToken func(string) (string, int, error)
	SwitchReady   func(string) error
	SendChat      func(string, string) error
	Operate       func(string, string) error
	Quit          func(string) error
	Ping          func(string) error
	GetData       func(int, string) ([]string, error)
}

func initRpcClient() {
	client.UseService(&s)
}
