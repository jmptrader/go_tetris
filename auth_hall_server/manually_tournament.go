// used for manually create a tournament game
package main

import (
	"net/http"

	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

var (
	tournamentServer = hprose.NewHttpService()
)

type (
	tournamentStub   struct{}
	tournamentFilter struct{}
)

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
