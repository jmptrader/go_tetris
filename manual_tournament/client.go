package main

import (
	"fmt"

	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

type ts struct {
	GetAll func() ([]string, error)
	Add    func(awardGold, awardSilver int, sponsor string) error
	Delete func() error
}

type filter struct{}

func (filter) InputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Decrypt(data, tournamentKey)
}

func (filter) OutputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Encrypt(data, tournamentKey)
}

var (
	client hprose.Client
	stub   = new(ts)
)

func initClient() {
	client = hprose.NewHttpClient(fmt.Sprintf("http://%s:%s/", authServerIp, authServerRpcPort))
	client.SetKeepAlive(true)
	client.SetFilter(filter{})
	client.UseService(&stub)
}
