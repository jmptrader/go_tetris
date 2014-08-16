package main

import (
	"fmt"
	"time"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

var normalHall = types.NewNormalHall()
var tournamentHall *types.TournamentHall

var hostFormat string

func constructHost(ip string) string { return fmt.Sprintf(hostFormat, ip) }

func initHall() {
	hostFormat = `%s:` + gameServerSocketPort
	log.Debug("the host format is %s", hostFormat)

	go releaseExpires()
	go createTournamentForever()
}

func releaseExpires() {
	defer utils.RecoverFromPanic("release expire tables panic: ", log.Critical, releaseExpires)
	for {
		for tid, _ := range normalHall.GetExpireTables() {
			tt := normalHall.GetTableById(tid)
			if tt == nil {
				continue
			}
			// inform game server
			ip := tt.GetIp()
			if err := clients.GetStub(ip).Delete(tid); err != nil {
				log.Warn("can not inform game server %v to delete table %v: %v", ip, tid, err)
			}
			// release the busy users in cache, including the observers and players
			users.SetFree(tt.GetAllUsers()...)
			// release the expire table also
			normalHall.ReleaseExpireTable(tid)
		}
		time.Sleep(5 * time.Second)
	}
}

func createTournamentForever() {
	defer utils.RecoverFromPanic("create tournament hall panic: ", log.Critical, createTournamentForever)
	for {
		time.Sleep(5 * time.Second)
		if tournamentHall == nil || tournamentHall.TournamentEnded() {
			n := nextTournaments.FlashGet()
			ip := clients.BestServer()
			if ip == "" {
				continue
			}
			tournamentHall = types.NewTournamentHall(n.numCandidate, n.awardGold, n.awardSilver, constructHost(ip), n.sponsor)
		}
	}
}
