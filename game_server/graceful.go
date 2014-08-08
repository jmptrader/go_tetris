/*
	gracefully exit the program when kill or C-c command is executed
*/
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogames/go_tetris/utils"
)

func initGraceful() { go notify() }

func notify() {
	defer utils.RecoverFromPanic("notify panic: ", log.Critical, notify)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigs:
	}
	deactivateServer(true)
	log.Info("gracefully exit the program...")
	time.Sleep(time.Second)
	os.Exit(1)
}
