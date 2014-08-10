// used for manually create a tournament game
package main

import (
	"container/list"
	"fmt"
	"net/http"

	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

const (
	defaultNumCandidate = 1 << 6
	defaultAwardGold    = 6
	defaultAwardSilver  = 4
	defaultSponsor      = "cointetris"
)

var (
	nextTournaments  = newNts()
	defaultNt        = newNt(defaultNumCandidate, defaultAwardGold, defaultAwardSilver, defaultSponsor)
	tournamentServer = hprose.NewHttpService()
)

type (
	nt struct {
		numCandidate           int
		awardGold, awardSilver int
		sponsor                string
	}
	nts                   struct{ *list.List }
	tournamentStub        struct{}
	tournamentFilter      struct{}
	tournamentServerEvent struct{}
)

func newNts() *nts { return &nts{List: list.New()} }

func newNt(numCandidate, awardGold, awardSilver int, sponsor string) nt {
	return nt{numCandidate: numCandidate, awardGold: awardGold, awardSilver: awardSilver, sponsor: sponsor}
}

func (n nt) String() string {
	return fmt.Sprintf("numCandidate: %d, award to gold: %d, to silver: %d and sponsor is %s\n",
		n.numCandidate, n.awardGold, n.awardSilver, n.sponsor)
}

// add new next tournament
func (ns *nts) Add(numCandidate, awardGold, awardSilver int, sponsor string) {
	for i := 1; i <= 10; i++ {
		if numCandidate == (1 << uint(i)) {
			goto add
		}
	}
	fmt.Errorf("the number of candidate should be the power of 2 and should be larger than 1")
add:
	ns.PushBack(newNt(numCandidate, awardGold, awardSilver, sponsor))
}

// delete all tournament
func (ns *nts) Delete() {
	for e := ns.Front(); e != nil; e = ns.Front() {
		ns.Remove(e)
	}
}

// get all next tournament
func (ns *nts) GetAll() []string {
	if ns.Len() == 0 {
		return nil
	}
	ss := make([]string, 0)
	for e := ns.Front(); e != nil; e = e.Next() {
		ss = append(ss, e.Value.(nt).String())
	}
	return ss
}

// flash get
func (ns *nts) FlashGet() nt {
	if ns.Len() <= 0 {
		return defaultNt
	}
	ret := ns.Front().Value.(nt)
	ns.Remove(ns.Front())
	return ret
}

// tournament stubs
func (tournamentStub) GetAll() []string { return nextTournaments.GetAll() }

func (tournamentStub) Add(numCandidate, awardGold, awardSilver int, sponsor string) {
	nextTournaments.Add(numCandidate, awardGold, awardSilver, sponsor)
}

func (tournamentStub) Delete() { nextTournaments.Delete() }

func (tournamentFilter) InputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Decrypt(data, tournamentKey)
}

func (tournamentFilter) OutputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Encrypt(data, tournamentKey)
}

func initTournamentServer() {
	tournamentServer.AddMethods(tournamentStub{})
	tournamentServer.SetFilter(tournamentFilter{})
	go serveTournament()
}

func serveTournament() {
	if err := http.ListenAndServe(":"+tournamentPort, tournamentServer); err != nil {
		panic(err)
	}
}
