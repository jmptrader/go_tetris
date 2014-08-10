package main

import "fmt"

func init() {
	initFlags()
	initConf()
	initClient()
}

func main() {
	fmt.Println("pass init")
}
