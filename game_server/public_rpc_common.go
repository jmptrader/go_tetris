/*
	socket server
*/
package main

import (
	"fmt"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/utils/queue"
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

// quit a game
func handleQuit(tid, uid int, nickname string, isOb, is1p, isTournament bool) {
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
		tableDatas.DeleteTable(tid)
		return
	}
	var msg string
	if isOb {
		msg = fmt.Sprintf("观战者 %s 退出房间", nickname)
	} else {
		msg = fmt.Sprintf("玩家 %s 退出房间", nickname)
	}
	tableDatas.SetData(tid, newResponse(descSysMsg, msg).toJson(), queue.BelongToAll)
}

// inform the client side to refresh the table information
func handleRefresh(tid int, isTournament bool) {
	if isTournament {
		tableDatas.SetData(tid, newResponse(descRefreshTournamentTableInfo, tid).toJson(), queue.BelongToAll)
	} else {
		tableDatas.SetData(tid, newResponse(descRefreshNormalTableInfo, tid).toJson(), queue.BelongToAll)
	}
}

// send sys msg
func handleSysMsg(tid int, msg string) {
	tableDatas.SetData(tid, newResponse(descSysMsg, msg).toJson(), queue.BelongToAll)
}

// send chat
func handleChat(tid int, isOb bool, msg string) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	if table.IsStart() && isOb {
		tableDatas.SetData(tid, newResponse(descChatMsg, msg).toJson(), queue.BelongToObs)
		return
	}
	tableDatas.SetData(tid, newResponse(descChatMsg, msg).toJson(), queue.BelongToAll)
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
	handleRefresh(tid, isTournament)
}

// handle operate
func handleOperate(tid int, is1p bool, op string) {
	table := tables.GetTableById(tid)
	if table == nil {
		return
	}
	if !table.IsStart() {
		return
	}
	var g *tetris.Game
	if is1p {
		g = table.GetGame1p()
	} else {
		g = table.GetGame2p()
	}
	switch op {
	case opDown:
		g.MoveDown()
	case opDrop:
		g.DropDown()
	case opLeft:
		g.MoveLeft()
	case opRight:
		g.MoveRight()
	case opRotate:
		g.Rotate()
	case opHold:
		g.Hold()
	default:
		log.Debug("unknown operation: %s\n", op)
	}
}

// inform the auth server, some one is going to ob a game
func obGame(tid, uid int, isTournament bool) error {
	if isTournament {
		return authServerStub.ObTournament(tid, uid)
	}
	return authServerStub.Join(tid, uid, true)
}
