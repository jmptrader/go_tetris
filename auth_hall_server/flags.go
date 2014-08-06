package main

import "flag"

var (
	confPath = flag.String("conf", "./default.conf", "path to configuration file")
	debug    = flag.Bool("debug", true, "debug enable")
)

func initFlags() { flag.Parse() }
