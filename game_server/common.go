package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

var errNilConn = fmt.Errorf("the connection is nil")

// receive data
func recv(user *types.User) (d requestData, err error) {
	conn := user.GetReadConn()
	if conn == nil {
		err = errNilConn
		return
	}
	b, err := utils.ReadDataOverTcp(conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &d)
	return
}

// receive data
func recvDefault(conn *net.TCPConn) (d requestData, err error) {
	if conn == nil {
		err = errNilConn
		return
	}
	b, err := utils.ReadDataOverTcp(conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &d)
	return
}

// send data
func send(user *types.User, desc string, data interface{}) error {
	conn := user.GetWriteConn()
	if conn == nil {
		log.Debug("the connection is nil, can not send")
		return errNilConn
	}
	err := utils.SendDataOverTcp(conn, newResponse(desc, data).toJson())
	if err != nil {
		log.Debug("can not send response data ->\ndesc: %v, data: %v, error: %v", desc, data, err)
	}
	return err
}

// send to all
func sendAll(desc string, val interface{}, conns ...*types.User) {
	for _, c := range conns {
		send(c, desc, val)
	}
}

// send data
func sendDefault(conn *net.TCPConn, desc string, data interface{}) error {
	if conn == nil {
		log.Debug("the connection is nil, can not send")
		return errNilConn
	}
	err := utils.SendDataOverTcp(conn, newResponse(desc, data).toJson())
	if err != nil {
		log.Debug("can not send response data ->\ndesc: %v, data: %v, error: %v", desc, data, err)
	}
	return err
}

// close a user's connection
func closeConn(users ...*types.User) {
	for _, u := range users {
		if err := u.Close(); err != nil {
			log.Debug("can not close a user's connection: %v", err)
		}
	}
}

// close a connection
func closeConnDefault(conns ...*net.TCPConn) {
	for _, conn := range conns {
		if conn == nil {
			continue
		}
		if err := conn.Close(); err != nil {
			log.Debug("can not close the connection: %v", err)
		}
	}
}

// response from game server to client
type responseData struct {
	Description string      `json:"desc"`
	Data        interface{} `json:"data"`
}

func newResponse(desc string, data interface{}) responseData {
	return responseData{Description: desc, Data: data}
}

func (r responseData) String() string {
	return fmt.Sprintf("\ndescription: %v, data: %v", r.Description, r.Data)
}

func (r responseData) toJson() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Debug("can not json marshal the data %v: %v", r, err)
	}
	return b
}

// request from client to game server
type requestData struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

func (d requestData) String() string {
	return fmt.Sprintf("\nrequest command: %s, data: %s", d.Cmd, d.Data)
}
