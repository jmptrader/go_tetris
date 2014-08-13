/*
	rpc server stubs
*/
package main

import (
	"fmt"
	"time"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/timer"
	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

// auth server inform game server to become inactive
// do not accept any new connections
// handle the current connections
func (stub) Deactivate() {
	go deactivateServer(false)
}

// count down after a game start
func countDown(table *types.Table) {
	t := timer.NewTimer(1000)
	t.Start()
	for i := 3; i > 0; i-- {
		sendAll(descStart, i, table.GetAllConns()...)
		t.Wait()
	}
	t.Stop()
	sendAll(descStart, 0, table.GetAllConns()...)
}

// auth server inform game server to start a table
func (stub) Start(tid int) {
	go func() {
		defer utils.RecoverFromPanic("game panic: ", log.Critical, nil)
		table := tables.GetTableById(tid)
		if table == nil {
			log.Critical("start the game but table is nil")
			return
		}
		countDown(table)
		table.StartGame()
		if !table.IsStart() {
			log.Debug("why the game is not start?")
		}
		go func() {
			defer utils.RecoverFromPanic("update timer panic: ", log.Critical, nil)
			table.UpdateTimer()
		}()
		serveGame(tid)
	}()
}

// create new table
func (stub) Create(tid int) error {
	return tables.NewTable(tid, "", "", -1)
}

// delete a table
func (stub) Delete(tid int) error {
	log.Info("auth server informs game server to delete the table %d", tid)
	t := tables.GetTableById(tid)
	if t == nil {
		err := fmt.Errorf("can not delete the table %d because the table is not exist.", tid)
		log.Critical("%v", err)
		return err
	}
	sendAll(descError, "桌子长时间不开始游戏, 或者由于其他原因, 桌子已经被取消.", t.GetAllConns()...)
	closeConn(t.GetAllConns()...)
	tables.DelTable(tid)
	return nil
}

// auth server inform game server the game result
func (stub) SetNormalGameResult(tid, winnerUid, bet int) {
	construct := func(win bool, bet int) (str string) {
		if win {
			str = "你很厉害哦!!"
			if bet > 0 {
				str += fmt.Sprintf(" 本局游戏你赢得了 %d mBTC", bet)
			}
			return str
		}
		if bet > 0 {
			str = fmt.Sprintf("本局游戏你输掉了 %d mBTC! ", bet)
		}
		str += "再接再厉!"
		return str
	}
	table := tables.GetTableById(tid)
	if table == nil {
		log.Critical("set normal game result but table is nil")
		return
	}
	switch winnerUid {
	case table.Get1pUid():
		send(table.Get1pConn(), descGameWin, construct(true, bet))
		send(table.Get2pConn(), descGameLose, construct(false, bet))
		sendAll(descGameResult, "1P 赢得本局游戏", table.GetObConns()...)
	case table.Get2pUid():
		send(table.Get2pConn(), descGameWin, construct(true, bet))
		send(table.Get1pConn(), descGameLose, construct(false, bet))
		sendAll(descGameResult, "2P 赢得本局游戏", table.GetObConns()...)
	default:
		log.Debug("the winner uid is neither 1p nor 2p, who is it: %v", winnerUid)
	}
	table.ResetTable()
	refreshTable(tid, false)
}

// TODO: not confirmed yet
func (stub) SetTournamentResult(tid, winnerUid int, isFinalRound bool) {
	construct := func(win, isFinalRound bool) (str string) {
		if win {
			if isFinalRound {
				str = "恭喜你获得冠军!"
			} else {
				str = "恭喜你获得进入下一轮游戏的资格!"
			}
		} else if isFinalRound {
			str = "恭喜你获得亚军!"
		} else {
			str = "不要气馁, 再接再厉!"
		}
		return
	}
	table := tables.GetTableById(tid)
	if table == nil {
		log.Critical("set tournament result but table is nil")
		return
	}
	switch winnerUid {
	case table.Get1pUid():
		send(table.Get1pConn(), descGameWin, construct(true, isFinalRound))
		send(table.Get2pConn(), descGameLose, construct(false, isFinalRound))
		closeConn(table.Get2pConn())
		sendAll(descGameResult, "1P 赢得本局游戏", table.GetObConns()...)
	case table.Get2pUid():
		send(table.Get2pConn(), descGameWin, construct(true, isFinalRound))
		send(table.Get1pConn(), descGameLose, construct(false, isFinalRound))
		closeConn(table.Get1pConn())
		sendAll(descGameResult, "2P 赢得本局游戏", table.GetObConns()...)
	default:
	}
	table.ResetTable()
	table.QuitAllObs()
}

// game server serve the game
func serveGame(tid int) {
	table := tables.GetTableById(tid)
	if table == nil {
		log.Critical("serve game but table is nil")
		return
	}
	for {
		select {

		// table timer
		case remain := <-table.RemainedSecondsChan:
			log.Debug("remain time in seconds: %d", remain)
			sendAll(descTimer, remain, table.GetAllConns()...)

		// game over
		case gameover := <-table.GameoverChan:
			log.Debug("table game over chan: %v", gameover)
			switch gameover {
			case types.GameoverNormal:
				// normal game over
				gameOver(tid)
			case types.Gameover1pQuit:
				// 1p quit, game over, 2p winner
				gameOver(tid, false)
			case types.Gameover2pQuit:
				// 2p quit, game over, 1p winner
				gameOver(tid, true)
			}
			return

		// 1p
		case msg := <-table.GetGame1p().MsgChan:
			log.Debug("1p msg: %v", msg)
			switch msg.Description {
			// ko, audio only send to the player himself
			case tetris.DescAudio, tetris.DescKo:
				sendAll(desc1p, msg, table.Get1pConn())
			// clear, combo, attack only sends to the player and obs
			case tetris.DescClear, tetris.DescCombo, tetris.DescAttack:
				sendAll(desc1p, msg, table.Get1pConn())
				sendAll(desc1p, msg, table.GetObConns()...)
			// the others send to all
			default:
				sendAll(desc1p, msg, table.GetAllConns()...)
			}

		case beingKo := <-table.GetGame1p().BeingKOChan:
			log.Debug("1p being ko: %v", beingKo)
			if beingKo {
				table.GetGame2p().KoOpponent()
				ko := table.GetGame2p().GetKo()
				sendAll(desc1p, tetris.NewMessage(tetris.DescBeingKo, ko), table.Get1pConn())
				sendAll(desc1p, tetris.NewMessage(tetris.DescBeingKo, ko), table.GetObConns()...)
				log.Debug("number of 2p ko: %d", ko)
				if ko >= 5 {
					log.Debug("send true to 1p gameover chan")
					table.GetGame1p().GameoverChan <- true
				}
			}

		// attack 2p
		case attack := <-table.GetGame1p().AttackChan:
			log.Debug("attacking 2p %d lines", attack)
			table.GetGame2p().BeingAttacked(attack)

		// 1p game over, 2p win
		case gameover := <-table.GetGame1p().GameoverChan:
			log.Debug("1p game over: %v", gameover)
			if gameover {
				gameOver(tid, false)
				return
			}

		// 2p
		case msg := <-table.GetGame2p().MsgChan:
			log.Debug("2p msg: %v", msg)
			// ko, audio only send to the player himself
			switch msg.Description {
			case tetris.DescAudio, tetris.DescKo:
				sendAll(desc2p, msg, table.Get2pConn())
			case tetris.DescClear, tetris.DescCombo, tetris.DescAttack:
				sendAll(desc2p, msg, table.Get2pConn())
				sendAll(desc2p, msg, table.GetObConns()...)
			default:
				sendAll(desc2p, msg, table.GetAllConns()...)
			}

		// attack 1p
		case attack := <-table.GetGame2p().AttackChan:
			log.Debug("attacking 1p %d lines", attack)
			table.GetGame1p().BeingAttacked(attack)

		// 2p game over, 1p win
		case gameover := <-table.GetGame2p().GameoverChan:
			log.Debug("2p game over: %v", gameover)
			if gameover {
				gameOver(tid, true)
				return
			}

		case beingKo := <-table.GetGame2p().BeingKOChan:
			log.Debug("2p being ko: %v", beingKo)
			if beingKo {
				table.GetGame1p().KoOpponent()
				ko := table.GetGame1p().GetKo()
				sendAll(desc2p, tetris.NewMessage(tetris.DescBeingKo, ko), table.Get2pConn())
				sendAll(desc2p, tetris.NewMessage(tetris.DescBeingKo, ko), table.GetObConns()...)
				log.Debug("number of 1p ko: %d", ko)
				if ko >= 5 {
					log.Debug("send true to 2p gameover chan")
					table.GetGame2p().GameoverChan <- true
				}
			}

		case <-time.After(time.Second * 2):
			log.Debug("do not receive any msg in 2 seconds: the game should be ended")
			return
		}
	}
}

// stop the game
// inform the auth server that the game is over
func gameOver(tid int, is1pWin ...bool) {
	log.Debug("table %d is game over, setting game result", tid)
	table := tables.GetTableById(tid)
	if table == nil {
		log.Critical("game over but the table is nil")
		return
	}
	table.StopGame()
	var is1pWinner = false
	var winner, loser int
	var err error

	// normal checker
	if len(is1pWin) == 0 {
		g1p, g2p := table.GetGame1p(), table.GetGame2p()
		switch {
		case g1p.GetKo() > g2p.GetKo():
			is1pWinner = true
		case g1p.GetKo() < g2p.GetKo():
		default:
			if g1p.GetScore() >= g2p.GetScore() {
				is1pWinner = true
			}
		}
	} else {
		is1pWinner = is1pWin[0]
	}

	// inform the auth server
	if is1pWinner {
		winner, loser = table.Get1pUid(), table.Get2pUid()
	} else {
		winner, loser = table.Get2pUid(), table.Get1pUid()
	}

	// 1e5 magic number
	if tid >= 1e5 {
		err = authServerStub.SetTournamentResult(tid, winner, loser)
	} else {
		err = authServerStub.SetNormalGameResult(tid, winner, loser)
	}
	if err != nil {
		log.Warn("can not set game result for table %d: %v", tid, err)
	}
}
