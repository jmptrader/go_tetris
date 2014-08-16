/*
	rpc server stubs
*/
package main

import (
	"fmt"
	"time"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
	"github.com/gogames/go_tetris/utils/queue"
)

const (
	sessKeyTid          = "tableId"
	sessKeyUid          = "userId"
	sessKeyNickname     = "nickname"
	sessKeyIsOb         = "isOb"
	sessKeyIsTournament = "isTournament"
	sessKeyIs1p         = "is1P"
)

var (
	getTidFromSession          = func(sessionId string) int { return session.GetSession(sessKeyTid, sessionId).(int) }
	getUidFromSession          = func(sessionId string) int { return session.GetSession(sessKeyUid, sessionId).(int) }
	getIsObFromSession         = func(sessionId string) bool { return session.GetSession(sessKeyIsOb, sessionId).(bool) }
	getIs1pFromSession         = func(sessionId string) bool { return session.GetSession(sessKeyIs1p, sessionId).(bool) }
	getIsTournamentFromSession = func(sessionId string) bool { return session.GetSession(sessKeyIsTournament, sessionId).(bool) }
	getNicknameFromSession     = func(sessionId string) string { return session.GetSession(sessKeyNickname, sessionId).(string) }
)

func (pubStub) ValidateToken(token string) (sessionId string, index int) {
	uid, nickname, isApply, isOb, isTournament, tid, err := utils.ParseToken(token)
	if err != nil {
		panic(err)
	}
	u := types.NewUser(uid, "", "", nickname, "")
	switch {
	case isApply:
		// apply for tournament
		tid, err := authServerStub.Apply(uid)
		if err != nil {
			log.Warn("can not apply for tournament, auth server error: %v", err)
			panic(fmt.Sprintf("报名失败, 错误: %v", err))
		}
		if !tables.IsTableExist(tid) {
			tables.NewTable(tid, "", "", 0)
		}
		// the err should always be nil actually
		if err := tables.JoinTable(tid, u, false); err != nil {
			log.Debug("can not join the table, game server error: %v", err)
			panic(fmt.Sprintf("无法加入桌子, 错误: %v", err))
		}
		handleRefresh(tid, true)
		handleSysMsg(tid, fmt.Sprintf("参赛者 %s 加入", nickname))
	case isOb:
		// inform the auth server that some one is going to observe a game
		if err := obGame(tid, uid, isTournament); err != nil {
			log.Warn("can not ob a game, auth server error: %v", err)
			panic(fmt.Sprintf("无法观战, 错误: %v", err))
		}
		if err := tables.JoinTable(tid, u, true); err != nil {
			log.Critical("can not ob a game, game server error: %v", err)
			panic(fmt.Sprintf("无法观战, 错误: %v", err))
		}
		// do not inform all people that an observer join the table
		// refreshTable(tid, isTournament)
		handleSysMsg(tid, fmt.Sprintf("用户 %s 进入观战", nickname))
	default:
		// normal hall
		if err := authServerStub.Join(tid, uid, false); err != nil {
			log.Warn("can not join a game, auth server error: %v", err)
			panic(fmt.Sprintf("无法加入桌子, 错误: %v", err))
		}
		if err := tables.JoinTable(tid, u, isOb); err != nil {
			log.Critical("can not join a game, game server error: %v", err)
			panic(fmt.Sprintf("无法加入桌子, 错误: %v", err))
		}
		handleRefresh(tid, false)
		handleSysMsg(tid, fmt.Sprintf("玩家 %s 加入游戏", nickname))
	}

	sessionId = session.CreateSession()
	session.SetSession(sessKeyUid, uid, sessionId)
	session.SetSession(sessKeyNickname, nickname, sessionId)
	session.SetSession(sessKeyIsOb, isOb, sessionId)
	session.SetSession(sessKeyIsTournament, isTournament, sessionId)
	session.SetSession(sessKeyTid, tid, sessionId)
	session.SetSession(sessKeyIs1p, tables.GetTableById(tid).Is1p(uid), sessionId)

	index = tableDatas.Index(tid)
	return
}

// switch ready state
func (pubStub) SwitchReady(sessionId string) {
	handleReady(getTidFromSession(sessionId),
		getUidFromSession(sessionId),
		getIsObFromSession(sessionId),
		getIsTournamentFromSession(sessionId))
}

// send chat info
func (pubStub) SendChat(msg string, sessionId string) {
	handleChat(getTidFromSession(sessionId),
		getIsObFromSession(sessionId),
		fmt.Sprintf("%s: %s", getNicknameFromSession(sessionId), msg))
}

// operate game
func (pubStub) Operate(op string, sessionId string) {
	if !getIsObFromSession(sessionId) {
		handleOperate(getTidFromSession(sessionId),
			getIs1pFromSession(sessionId),
			op)
	}
}

// quit
func (pubStub) Quit(sessionId string) {
	handleQuit(getTidFromSession(sessionId),
		getUidFromSession(sessionId),
		getNicknameFromSession(sessionId),
		getIsObFromSession(sessionId),
		getIs1pFromSession(sessionId),
		getIsTournamentFromSession(sessionId))

	session.DelSession(sessionId)
}

// ping
func (pubStub) Ping(sessionId string) {}

// get data
func (pubStub) GetData(index int, sessionId string) (res []interface{}, newIndex int) {
	tid := getTidFromSession(sessionId)
	var belong = queue.BelongToObs
	if !getIsObFromSession(sessionId) {
		if getIs1pFromSession(sessionId) {
			belong = queue.BelongTo1p
		} else {
			belong = queue.BelongTo2p
		}
	}
	var count = 200
	for count > 0 {
		count--
		if res = tableDatas.GetData(tid, index, belong); res != nil {
			newIndex = tableDatas.Index(tid)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	panic("no new data")
}
