/*
	socket server listen on 843 for AS3 requesting for pocily file
*/
package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gogames/go_tetris/utils"
)

func initPolicyFileSocketServer() {
	socketPolicyFile = []byte(fmt.Sprintf(`
	<?xml version="1.0" encoding="UTF-8"?>
	
	<cross-domain-policy>
		<allow-access-from domain="*.cointetris.com" to-ports="%s" />
	</cross-domain-policy>`, gameServerSockPort))
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":843")
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

	log.Info("successfully initialization, as3 policy file socket server accepting connection...")
	go func() {
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.Critical("do not accept tcp connection to policy file socket server: %v", err)
				continue
			}
			go servePolicyFileRequest(conn)
		}
	}()
}

const bufPFR = 1 << 5

var socketPolicyFile []byte

func servePolicyFileRequest(conn *net.TCPConn) {
	defer utils.RecoverFromPanic("serve policy file request tcp connection panic: ", log.Critical, nil)
	defer conn.Close()
	bufprf := make([]byte, bufPFR)
	n, err := conn.Read(bufprf)
	if err != nil {
		log.Debug("policy file server can not read from tcp connection: %v\nthe length of n is %d\n", err, n)
		return
	}
	equal := strings.TrimSpace(fmt.Sprintf("%s", bufprf[:22])) == "<policy-file-request/>"
	if equal {
		_, err := conn.Write(socketPolicyFile)
		if err != nil {
			log.Debug("can not send policy file: %v\n%s\n", err, socketPolicyFile)
			return
		}
		log.Debug("successfully send policy file: %s", socketPolicyFile)
		return
	}
	log.Debug("the string is %s, equal ? %v", bufprf, equal)
}
