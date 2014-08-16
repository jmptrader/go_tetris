package main

import "github.com/gogames/go_tetris/utils"

var session = utils.NewSessionStore(10)

const garbageBuffer = 1 << 10

func initSession() {
	session.EnableGarbageChan(garbageBuffer)
	go gc()
}

func gc() {
	for {
		sess := <-session.GarbageSession
		handleQuit(sess.Get(sessKeyTid).(int),
			sess.Get(sessKeyUid).(int),
			sess.Get(sessKeyNickname).(string),
			sess.Get(sessKeyIsOb).(bool),
			sess.Get(sessKeyIs1p).(bool),
			sess.Get(sessKeyIsTournament).(bool))
	}
}
