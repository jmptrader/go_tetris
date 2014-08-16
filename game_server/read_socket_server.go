/*
	socket server
*/
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

// server read data from client side over tcp
// Client send
// Server receive
func serveRead(conn *net.TCPConn, token string) {
	// parse the token, see what to do next
	uid, nickname, isApply, isOb, isTournament, tid, err := utils.ParseToken(token)
	if err != nil {
		log.Debug("can not parse the token: %v", err)
		closeConnDefault(conn)
		return
	}
	// create a new user, add it into tables
	u := types.NewUser(uid, "", "", nickname, "")
	u.SetReadConn(conn)
	switch {
	case isApply:
		// apply for tournament
		tid, err := authServerStub.Apply(uid)
		if err != nil {
			log.Warn("can not apply for tournament, auth server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("报名失败, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		if !tables.IsTableExist(tid) {
			tables.NewTable(tid, "", "", 0)
		}
		// the err should always be nil actually
		if err := tables.JoinTable(tid, u, false); err != nil {
			log.Debug("can not join the table, game server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		refreshTable(tid, true)
		sendSysMsg(tid, fmt.Sprintf("参赛者 %s 加入", nickname))
	case isOb:
		// inform the auth server that some one is going to observe a game
		if err := obGame(tid, uid, isTournament); err != nil {
			log.Warn("can not ob a game, auth server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("无法观战, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		if err := tables.JoinTable(tid, u, true); err != nil {
			log.Critical("can not ob a game, game server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("无法观战, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		// do not inform all people that an observer join the table
		// refreshTable(tid, isTournament)
		sendSysMsg(tid, fmt.Sprintf("用户 %s 进入观战", nickname))
	default:
		// normal hall
		if err := authServerStub.Join(tid, uid, false); err != nil {
			log.Warn("can not join a game, auth server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		if err := tables.JoinTable(tid, u, isOb); err != nil {
			log.Critical("can not join a game, game server error: %v", err)
			sendDefault(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConnDefault(conn)
			return
		}
		refreshTable(tid, false)
		sendSysMsg(tid, fmt.Sprintf("玩家 %s 加入游戏", nickname))
	}
	handleRead(u, uid, tid, nickname, isOb, tables.GetTableById(tid).Is1p(uid), isTournament)
}

func handleRead(conn *types.User, uid, tid int, nickname string, isOb, is1p, isTournament bool) {
	var handleQuit = func() {
		quit(tid, uid, nickname, isOb, is1p, isTournament)
		closeConn(conn)
		refreshTable(tid, isTournament)
	}
	// check in case the write connection does not join
	var exceptionChecker = func() {
		defer utils.RecoverFromPanic("exception checker panic: ", log.Critical, handleQuit)
		t := time.Now()
		for {
			if time.Since(t).Seconds() >= float64(writeConnJoinWindow) {
				handleQuit()
				return
			}
			if conn.GetWriteConn() != nil {
				return
			}
			time.Sleep(time.Second)
		}
	}
	defer utils.RecoverFromPanic("handle connection panic: ", log.Critical, handleQuit)
	if err := sendDefault(conn.GetReadConn(), descAuthSuccess, 0); err != nil {
		log.Debug("can not send auth success info: %v", err)
		handleQuit()
		return
	}
	go exceptionChecker()
	table := tables.GetTableById(tid)
forLoop:
	for {
		if table == nil {
			log.Debug("the table is already been deleted")
			handleQuit()
			return
		}

		if err := conn.SetReadTimeoutInSecs(pingDuration); err != nil {
			log.Debug("can not set heartbeat deadline, error: %v", err)
			handleQuit()
			return
		}
		// receive data from client
		data, err := recv(conn)
		if err != nil {
			log.Debug("can not receive request from table %d, user %s: %v", tid, nickname, err)
			handleQuit()
			return
		}
		switch data.Cmd {
		case cmdChat:
			sendChat(tid, isOb, fmt.Sprintf("%s: %s", nickname, data.Data))
		case cmdReady:
			handleReady(tid, uid, isOb, isTournament)
		case cmdQuit:
			// quit a game
			handleQuit()
			return
		case cmdOperate:
			if !table.IsStart() {
				continue forLoop
			}
			var g *tetris.Game
			if is1p {
				g = table.GetGame1p()
			} else {
				g = table.GetGame2p()
			}
			switch data.Data {
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
			}
		case cmdPing:
			// get ping from client
			log.Debug("get ping from %s", conn.GetReadConn().RemoteAddr().String())
		default:
			log.Debug("get strange packet from client: %+v", data)
		}
	}
}

// inform the auth server, some one is going to ob a game
func obGame(tid, uid int, isTournament bool) error {
	if isTournament {
		return authServerStub.ObTournament(tid, uid)
	}
	return authServerStub.Join(tid, uid, true)
}
