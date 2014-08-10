package main

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/config"
	"github.com/gogames/go_tetris/utils"
)

var (
	conf config.ConfigContainer

	authServerIp      string
	authServerRpcPort string
	tournamentKey     []byte
)

func initConf() {
	var err error
	conf, err = config.NewConfig("json", *confPath)
	if err != nil {
		fmt.Printf("can not read configuration: %v", err)
		os.Exit(1)
	}

	authServerIp = conf.String("authServerIp")
	authServerRpcPort = conf.String("authServerRpcPort")
	tournamentKeyString := conf.String("tournamentKey")

	utils.CheckEmptyConf(authServerIp, authServerRpcPort, tournamentKeyString)

	tournamentKey = []byte(tournamentKeyString)
}
