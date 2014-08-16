package main

import (
	"time"

	"github.com/gogames/go_tetris/utils/queue"
)

var tableDatas = queue.NewTableDatas()

func initTableDatas() {
	go printTDForDebug()
}

func printTDForDebug() {
	for {
		log.Debug(tableDatas.PrintForDebug())
		time.Sleep(10 * time.Second)
	}
}
