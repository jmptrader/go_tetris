package main

import "flag"

var confPath = flag.String("conf", "./default.conf", "configuration of tournament controller")

func initFlags() { flag.Parse() }
