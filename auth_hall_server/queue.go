package main

import "time"

var funcQueue = make(chan func(), 1<<10)

func initQueue() {
	go execFuncs()
	log.Info("initialize the function queue...")
}

func pushFunc(f func()) {
	funcQueue <- f
}

func getFunc() func() {
	return <-funcQueue
}

func execFuncs() {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("queue panic: %v", err)
		}
		go execFuncs()
	}()
	for {
		select {
		case f := <-funcQueue:
			f()
		case <-time.After(10 * time.Second):
			// no function to be execute in 30 seconds
			if allGSReleased {
				progCanExit = true
			}
		}
	}
}
