package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/gogames/go_tetris/utils"
)

func handleConn(host, token string) {
	raddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		fmt.Println("can not resolve address:", err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Println("can not dial tcp:", err)
		return
	}
	defer conn.Close()

	// another goroutine handling read data
	go handleRead(conn)

	// authenticate
	if err := send(conn, cmdAuth, token); err != nil {
		fmt.Println("can not send data: ", err)
		return
	}

	go heartBeat(conn)

	// switch ready state
	if err := send(conn, cmdReady, "I am hacker~"); err != nil {
		fmt.Println("can not switch ready state: ", err)
		return
	}

	// send chat msg
	if err := send(conn, cmdChat, "hello all guys, sending chat msg"); err != nil {
		fmt.Println("can not send chat msg: ", err)
		return
	}

	// operate
	if err := send(conn, cmdOperate, opLeft); err != nil {
		fmt.Println("can not operate: ", err)
		return
	}

	// close the connection in 3 minute
	time.Sleep(time.Hour)
}

func handleRead(conn *net.TCPConn) {
	defer conn.Close()
	for {
		r, err := recv(conn)
		if err != nil {
			fmt.Println("can not read from connection: ", err)
			return
		}
		fmt.Printf("receive data: %+v\n", r)
	}
}

func heartBeat(conn *net.TCPConn) {
	defer conn.Close()
	for {
		if err := send(conn, cmdPing, ""); err != nil {
			fmt.Println("can not ping:", err)
			return
		}
		time.Sleep(2 * time.Second)
	}
}

// request to socket server
type requestData struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

func newRequestData(cmd, data string) requestData { return requestData{Cmd: cmd, Data: data} }

func (r requestData) ToJson() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		panic("can not json marshal the data: " + err.Error())
	}
	return b
}

// response from socket server
type responseData struct {
	Desc string      `json:"desc"`
	Data interface{} `json:"data"`
}

// send to server
func send(conn *net.TCPConn, cmd, data string) error {
	return utils.SendDataOverTcp(conn, newRequestData(cmd, data).ToJson())
}

// read from server
func recv(conn *net.TCPConn) (*responseData, error) {
	b, err := utils.ReadDataOverTcp(conn)
	if err != nil {
		return nil, err
	}
	r := new(responseData)
	err = json.Unmarshal(b, r)
	return r, err
}

const (
	opRotate = "rotate"
	opLeft   = "left"
	opRight  = "right"
	opDown   = "down"
	opDrop   = "drop"
	opHold   = "hold"
)

// request command
const (
	cmdAuth    = "auth"
	cmdPing    = "ping"
	cmdChat    = "chat"
	cmdOperate = "operate"
	cmdReady   = "switchState"
	cmdQuit    = "quit"
)

// response description
const (
	descError                      = "error"
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
