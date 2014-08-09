package main

import "github.com/gogames/go_tetris/utils"

// session store
var session = utils.NewSessionStore()

func initSession() {
	log.Info("successfully init session...")
	session.Init(querySessions())
	deleteSessions()
	log.Info("all sessions: %v", session)
}
