package main

func init() {
	initFlags()
	initConf()
	initLogger()
	initRpcClient()
	initServerStatus()
	initPolicyFileSocketServer()
	initSocketServer()
	initRpcServer()
	initTables()
	initGraceful()
}

func main() {
	c := make(chan bool)
	<-c
}
