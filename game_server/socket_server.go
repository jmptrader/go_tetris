/*
	socket server
*/
package main

import (
	"net"
	"os"
	"time"

	"github.com/gogames/go_tetris/utils"
)

func initSocketServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+gameServerSockPort)
	if err != nil {
		log.Critical("can not resolve tcp address: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Critical("can not listen tcp: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}

	log.Info("successfully initialization, socket server accepting connection...")
	go func() {
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.Critical("do not accept tcp connection: %v", err)
				continue
			}
			if !isServerActive() {
				log.Info("the game server is closing, do not accept new connections...")
				closeConnDefault(conn)
				continue
			}
			go serveAuth(conn)
		}
	}()
}

// request command
const (
	cmdAuthRead  = "authWrite"
	cmdAuthWrite = "authRead"
	cmdChat      = "chat"
	cmdOperate   = "operate"
	cmdReady     = "switchState"
	cmdQuit      = "quit"
	cmdPing      = "ping"
)

var (
	pingDuration        = 5
	authenDuration      = 10 * time.Second
	writeConnJoinWindow = 10
)

func serveAuth(conn *net.TCPConn) {
	defer utils.RecoverFromPanic("serve read tcp connection panic: ", log.Critical, nil)
	// keep the connection alive
	if err := conn.SetKeepAlive(true); err != nil {
		log.Debug("set keep alive error: %v", err)
	}
	// authenticate in 10 seconds
	if err := conn.SetReadDeadline(time.Now().Add(authenDuration)); err != nil {
		log.Debug("set authentication deadline error: %v", err)
	}
	// auth -> check the connection
	data, err := recvDefault(conn)
	if err != nil {
		log.Info("can not read from the tcp connection: %v", err)
		closeConnDefault(conn)
		return
	}
	switch data.Cmd {
	case cmdAuthRead:
		serveRead(conn, data.Data)
	case cmdAuthWrite:
		serveWrite(conn, data.Data)
	default:
		log.Debug("the first command is not auth read or write, the data is %v", data)
		closeConnDefault(conn)
	}
}
