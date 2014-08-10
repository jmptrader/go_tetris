/*
	tables store the information of tables, connections
*/
package main

import (
	"time"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

var tables = types.NewTables()

func initTables() {
	go printTablesForDebug()
}

func printTablesForDebug() {
	defer utils.RecoverFromPanic("print tables for debug panic: ", log.Critical, printTablesForDebug)
	for {
		time.Sleep(time.Minute)
		log.Debug("%s", tables.String())
	}
}
