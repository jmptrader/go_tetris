/*
	socket server
*/
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

const (
	opRotate = "rotate"
	opLeft   = "left"
	opRight  = "right"
	opDown   = "down"
	opDrop   = "drop"
	opHold   = "hold"
)

// response description
const (
	descAuthSuccess                = "authSuccess"
	descError                      = "error"
	descPing                       = "ping"
	descChatMsg                    = "chat"
	descRefreshNormalTableInfo     = "refreshNormal"
	descRefreshTournamentTableInfo = "refreshTournament"
	descSysMsg                     = "sysMsg"
	descStart                      = "start"
	desc1p                         = "1p"
	desc2p                         = "2p"
	descTimer                      = "timer"
	descGameWin                    = "win"
	descGameLose                   = "lose"
	descGameResult                 = "result"
)

func serveWrite(conn *net.TCPConn, token string) {
	// parse the token, see what to do next
	uid, _, _, _, _, tid, err := utils.ParseToken(token)
	if err != nil {
		log.Debug("can not parse the token: %v", err)
		closeConnDefault(conn)
		return
	}
	// create a new user, add it into tables
	table := tables.GetTableById(tid)
	if table == nil {
		log.Debug("the table is nil, can not join write connection")
		closeConnDefault(conn)
		return
	}
	if !table.IsUserExist(uid) {
		log.Debug("the user is not exist, can not join write connection")
		closeConnDefault(conn)
		return
	}
	u := table.GetUserById(uid)
	if u == nil {
		log.Debug("the user %d is nil, can not join write connection", uid)
		closeConnDefault(conn)
		return
	}
	u.SetWriteConn(conn)
	go heartBeat(u)
}

func heartBeat(u *types.User) {
	for {
		if err := send(u, descPing, time.Now().Unix()); err != nil {
			log.Debug("write tcp connection can not ping: %v", err)
			u.Close()
			return
		}
		time.Sleep(5 * time.Second)
	}
}

// quit a game
func quit(tid, uid int, nickname string, isOb, is1p, isTournament bool) {
	log.Debug("user %s quit the table %d", nickname, tid)
	if err := authServerStub.Quit(tid, uid, isTournament); err != nil {
		log.Warn("hprose error, can not quit user %s from table %d: %v", nickname, tid, err)
	}
	table := tables.GetTableById(tid)
	if table == nil {
		log.Debug("why the table %d is nil but also quit?", tid)
		return
	}
	if !isOb {
		if table.IsStart() {
			if is1p {
				gameOver(tid, false)
				// table.GameoverChan <- types.Gameover1pQuit
			} else {
				gameOver(tid, true)
				// table.GameoverChan <- types.Gameover2pQuit
			}
		}
	}
	table.Quit(uid)
	if table.HasNoPlayer() {
		tables.DelTable(tid)
		return
	}
	var msg string
	if isOb {
		msg = fmt.Sprintf("观战者 %s 退出房间", nickname)
	} else {
		msg = fmt.Sprintf("玩家 %s 退出房间", nickname)
	}
	sendAll(descSysMsg, msg, table.GetAllConns()...)
}

// inform the client side to refresh the table information
func refreshTable(tid int, isTournament bool) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	if isTournament {
		sendAll(descRefreshTournamentTableInfo, tid, table.GetAllConns()...)
	} else {
		sendAll(descRefreshNormalTableInfo, tid, table.GetAllConns()...)
	}
}

// send sys msg
func sendSysMsg(tid int, msg string) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	sendAll(descSysMsg, msg, table.GetAllConns()...)
}

// send chat
func sendChat(tid int, isOb bool, msg string) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	if table.IsStart() && isOb {
		sendAll(descChatMsg, msg, table.GetObConns()...)
		return
	}
	sendAll(descChatMsg, msg, table.GetAllConns()...)
}

// handle ready
func handleReady(tid, uid int, isOb, isTournament bool) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	if table.IsStart() {
		log.Debug("receive a ready command after game is start: %v", tid)
		return
	}
	if isOb {
		log.Debug("observer can not switch ready state")
		return
	}
	if err := authServerStub.SwitchReady(tid, uid); err != nil {
		log.Warn("can not switch user's ready state: %v", err)
		return
	}
	refreshTable(tid, isTournament)
}
