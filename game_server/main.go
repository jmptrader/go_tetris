package main

func init() {
	initFlags()
	initConf()
	initLogger()
	initRpcClient()
	initServerStatus()
	initRpcServer()
	initPubRpcServer()
	initTables()
	initTableDatas()
	initSession()
	initGraceful()
}

func main() {
	c := make(chan bool)
	<-c
}
