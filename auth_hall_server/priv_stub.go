package main

import (
	"fmt"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

var errInsufficientEnergy = fmt.Errorf("能量不足, 每局游戏需消耗 1 能量")

// register a game server
func (privStub) Register(maxConn int, ctx interface{}) {
	clients.NewGameServer(utils.GetIp(ctx), maxConn)
}

// deactivate a game server
func (privStub) Deactivate(ctx interface{}) {
	clients.Deactivate(utils.GetIp(ctx))
}

// unregister a game server
func (privStub) Unregister(ctx interface{}) {
	clients.Delete(utils.GetIp(ctx))
}

// join a game
func (privStub) Join(tid, uid int, isOb bool) {
	t := normalHall.GetTableById(tid)
	u := getUserById(uid)
	if t == nil {
		panic(fmt.Errorf(errTableNotExist, tid))
	}
	if u == nil {
		panic(fmt.Errorf(errUserNotExist, uid))
	}
	// check busy
	if users.IsBusyUser(uid) {
		panic(errAlreadyInGame)
	}
	// check energy
	if u.GetEnergy() <= 0 {
		panic(errInsufficientEnergy)
	}
	// check balance
	if bet := t.GetBet(); u.GetBalance() < bet {
		panic(errBalNotSufficient)
	}
	if err := normalHall.JoinTable(tid, u, isOb); err != nil {
		panic(err)
	}
	// !!!! NOT UPDATE HERE, SHOULD UPDATE ON GAME START
	// SEE tableChecker
	// update energy, balance, freezed
	// if err := u.Update(types.NewUpdateInt(types.UF_Energy, u.GetEnergy()-1),
	// 	types.NewUpdateInt(types.UF_Balance, u.GetBalance()-bet),
	// 	types.NewUpdateInt(types.UF_Freezed, u.GetFreezed()+bet)); err != nil {
	// 	panic(err)
	// }
	// pushFunc(func() { insertOrUpdateUser(u) })
	users.SetBusy(uid)
}

// observe a tournament
func (privStub) ObTournament(tid, uid int) {
	u := getUserById(uid)
	if u == nil {
		panic(fmt.Errorf(errUserNotExist, uid))
	}
	if err := tournamentHall.JoinTable(tid, u, true); err != nil {
		panic(err)
	}
	users.SetBusy(uid)
}

// set normal game result
func (privStub) SetNormalGameResult(tid, winner, loser int, ctx interface{}) {
	t := normalHall.GetTableById(tid)
	t.Stop()

	// update winner info
	func() {
		w := getUserById(winner)
		upts := make([]types.UpdateInterface, 0)
		upts = append(upts, types.NewUpdateInt(types.UF_Balance, w.GetBalance()+t.GetBet()*2))
		upts = append(upts, types.NewUpdateInt(types.UF_Freezed, w.GetFreezed()-t.GetBet()))
		upts = append(upts, types.NewUpdateInt(types.UF_Win, w.Win+1))
		if (w.Win + 1) > (w.Level * w.Level) {
			upts = append(upts, types.NewUpdateInt(types.UF_Level, w.Level+1))
		}
		if err := w.Update(upts...); err != nil {
			log.Critical("set normal hall result, can not update winner %v: %v", w.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(w) })
	}()

	// update loser info
	func() {
		l := getUserById(loser)
		if err := l.Update(types.NewUpdateInt(types.UF_Freezed, l.GetFreezed()-t.GetBet()),
			types.NewUpdateInt(types.UF_Lose, l.Lose+1)); err != nil {
			log.Critical("set normal hall game result, can not update loser %v: %v", l.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(l) })
	}()

	// update busy timestamp
	users.SetBusy(t.GetAllUsers()...)

	if err := clients.GetStub(utils.GetIp(ctx)).SetNormalGameResult(tid, winner, t.GetBet()); err != nil {
		log.Warn("can not inform game server to set the game result: %v", err)
	}
}

// set tournament game result
func (privStub) SetTournamentResult(tid, winner, loser int) int {
	t := tournamentHall.GetTableById(tid)
	// update winner info
	w := getUserById(winner)
	func() {
		upts := make([]types.UpdateInterface, 0)
		upts = append(upts, types.NewUpdateInt(types.UF_Win, w.Win+1))
		if (w.Win + 1) > (w.Level * w.Level) {
			upts = append(upts, types.NewUpdateInt(types.UF_Level, w.Level+1))
		}
		if err := w.Update(upts...); err != nil {
			log.Critical("tournament hall -> can not update winner %v: %v", w.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(w) })
	}()

	// update loser info
	func() {
		l := getUserById(loser)
		if err := l.Update(types.NewUpdateInt(types.UF_Lose, l.Lose+1)); err != nil {
			log.Critical("tournament hall -> can not update loser %v: %v", l.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(l) })
	}()

	// update tournament hall
	tournamentHall.SetWinnerLoser(tid, winner)
	nid, err := tournamentHall.Allocate(w)
	if err != nil {
		log.Critical("tournament hall -> can not allocate user %v to next table: %v", w.Nickname, err)
		panic(err)
	}

	// winner continue to play, update busy timestamp
	users.SetBusy(winner)

	// loser and observers quit, set free
	users.SetFree(loser)
	users.SetFree(t.GetObservers()...)
	return nid
}

// TODO:
// apply for tournament
func (privStub) Apply(uid int) int {
	tid, err := tournamentHall.Apply(getUserById(uid))
	if err != nil {
		panic(err)
	}
	users.SetBusy(uid)
	return tid
}

// deprecated, useless, because SetTournamentResult already handle the Allocate
// allocate for tournament
// func (privStub) Allocate(uid int) int {
// 	tid, err := tournamentHall.Allocate(getUserById(uid))
// 	if err != nil {
// 		panic(err)
// 	}
// 	users.SetBusy(uid)
// 	return tid
// }

// quit a user
func (privStub) Quit(tid, uid int, isTournament bool) {
	if isTournament {
		t := tournamentHall.GetTableById(tid)
		if t == nil {
			log.Debug("why the table %d is nil but also quit? because of gracefully?", tid)
			return
		}
		t.Quit(uid)
	} else {
		t := normalHall.GetTableById(tid)
		if t == nil {
			log.Debug("why the table %d is nil but also quit? because of gracefully?", tid)
			return
		}
		t.Quit(uid)
		if t.HasNoPlayer() {
			normalHall.DelTable(tid)
		}
	}
	users.SetFree(uid)
}

var errTournamentDefaultReady = fmt.Errorf("tournament default is ready and can not set to not ready, what is wrong?")

// switch ready state
func (privStub) SwitchReady(tid, uid int, ctx interface{}) {
	if isTournament(tid) {
		panic(errTournamentDefaultReady)
	}
	t := normalHall.GetTableById(tid)
	u := getUserById(uid)
	if u == nil {
		log.Debug("the user %d is not exist", uid)
		panic(fmt.Errorf(errUserNotExist, uid))
	}
	if u.GetEnergy() <= 0 {
		panic(errInsufficientEnergy)
	}
	if u.GetBalance() < t.GetBet() {
		panic(errBalNotSufficient)
	}
	t.SwitchReady(uid)
	tableChecker(t, ctx)
}

func isTournament(tid int) bool {
	return tid >= 1e5
}

// check if the table should start
func tableChecker(t *types.Table, ctx interface{}) {
	if !t.ShouldStart() {
		return
	}
	if err := clients.GetStub(utils.GetIp(ctx)).Start(t.TId); err != nil {
		log.Critical("can not inform game server to start the table %d", t.TId)
		return
	}
	t.Start()

	u1p := getUserById(t.Get1pUid())
	u2p := getUserById(t.Get2pUid())
	if u1p == nil || u2p == nil {
		log.Critical("nil user? logic error")
		return
	}

	e1p := u1p.GetEnergy()
	e2p := u2p.GetEnergy()
	if e1p < 0 || e2p < 0 {
		log.Critical("negative energy\n1p %v has %v energy and 2p %v has %v energy", u1p.Nickname, e1p, u2p.Nickname, e2p)
	}
	if err := u1p.Update(types.NewUpdateInt(types.UF_Energy, e1p-1)); err != nil {
		log.Critical("can not update energy: %v", err)
	}
	if err := u2p.Update(types.NewUpdateInt(types.UF_Energy, e2p-1)); err != nil {
		log.Critical("can not update energy: %v", err)
	}

	// update balance
	if bet := t.GetBet(); bet > 0 {
		b1p, b2p := u1p.GetBalance(), u2p.GetBalance()
		f1p, f2p := u1p.GetFreezed(), u2p.GetFreezed()
		if b1p < bet || b2p < bet {
			log.Critical("balance is smaller than bet\nbet: %v, balance of 1p: %v, balance of 2p: %v", bet, b1p, b2p)
		}
		if f1p != 0 || f2p != 0 {
			log.Critical("freezed is not 0: user %v has %v freezed, user %v has %v freezed", u1p.Nickname, f1p, u2p.Nickname, f2p)
		}
		if err := u1p.Update(types.NewUpdateInt(types.UF_Balance, b1p-bet),
			types.NewUpdateInt(types.UF_Freezed, f1p+bet)); err != nil {
			log.Critical("can not update freezed and balance: %v", err)
		}
		if err := u2p.Update(types.NewUpdateInt(types.UF_Balance, b2p-bet),
			types.NewUpdateInt(types.UF_Freezed, f2p+bet)); err != nil {
			log.Critical("can not update freezed and balance: %v", err)
		}
	}

	// update userinfo in database
	pushFunc(func() { insertOrUpdateUser(u1p, u2p) })
}
