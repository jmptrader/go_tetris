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
	initGraceful()
}

func main() {
	c := make(chan bool)
	<-c
}
