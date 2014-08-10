package main

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/config"
	"github.com/gogames/go_tetris/utils"
)

var (
	conf                                                config.ConfigContainer
	btcUser, btcPass, btcServer                         string
	gameServerRpcPort, gameServerSocketPort             string
	dsn                                                 string
	logPath                                             string
	privRpcPort, pubRpcPort, tournamentPort             string
	emailId, emailUser, emailPass, emailHost, emailFrom string
	cookieEncryptKey, tokenEncryptKey, scryptSalt       string
	emailSMTPPort                                       int
	cookieDomain, crossDomainFile                       string
	privKey, tournamentKey                              []byte
)

func initConf() {
	var err error
	conf, err = config.NewConfig("json", *confPath)
	if err != nil {
		fmt.Printf("can not read configuration %v", err)
		os.Exit(1)
	}

	btcUser = conf.String("btcUser")
	btcPass = conf.String("btcPass")
	btcServer = conf.String("btcServer")
	gameServerRpcPort = conf.String("gameServerRpcPort")
	gameServerSocketPort = conf.String("gameServerSocketPort")
	dbUser := conf.String("dbUser")
	dbPass := conf.String("dbPass")
	dbProtocol := conf.String("dbProtocol")
	dbSockAddress := conf.String("dbSockAddress")
	dbName := conf.String("dbName")
	logPath = conf.String("log")
	privRpcPort = conf.String("privServerRpcPort")
	pubRpcPort = conf.String("publicServerRpcPort")
	tournamentPort = conf.String("tournamentPort")
	privKeyString := conf.String("privKey")
	tournamentKeyString := conf.String("tournamentKey")
	emailId = conf.String("emailIdentity")
	emailUser = conf.String("emailUsername")
	emailPass = conf.String("emailPassword")
	emailHost = conf.String("emailHost")
	emailFrom = conf.String("emailFrom")
	cookieEncryptKey = conf.String("cookieEncryptKey")
	tokenEncryptKey = conf.String("tokenEncryptKey")
	scryptSalt = conf.String("scryptSalt")
	cookieDomain = conf.String("domain")
	crossDomainFile = conf.String("crossDomainFile")
	emailSMTPPort, err = conf.Int("emailPort")
	if err != nil {
		panic("can not parse email smtp port: " + err.Error())
	}

	utils.CheckEmptyConf(btcUser, btcPass, btcServer, gameServerRpcPort, gameServerSocketPort, tournamentPort,
		dbUser, dbPass, dbProtocol, dbSockAddress, dbName,
		logPath, privRpcPort, pubRpcPort, privKeyString, tournamentKeyString,
		emailId, emailFrom, emailHost, emailPass, emailUser, emailSMTPPort,
		cookieEncryptKey, tokenEncryptKey, scryptSalt, crossDomainFile)

	// done configuration checking
	// initialize
	dsn = fmt.Sprintf("%s:%s@%s(%s)/%s", dbUser, dbPass, dbProtocol, dbSockAddress, dbName)
	utils.SetCookieKey([]byte(cookieEncryptKey))
	utils.SetTokenKey([]byte(tokenEncryptKey))
	utils.SetScryptSalt([]byte(scryptSalt))
	utils.SetEmailConf(emailId, emailUser, emailPass, emailHost, emailFrom, emailSMTPPort)
	if cookieDomain != "" {
		utils.SetDomain(cookieDomain)
	}
	privKey = []byte(privKeyString)
	tournamentKey = []byte(tournamentKeyString)
}
