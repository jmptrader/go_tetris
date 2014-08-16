package main

import (
	"fmt"
	"strconv"

	"github.com/gogames/go_tetris/types"
	"github.com/hprose/hprose-go/hprose"
)

var (
	client = hprose.NewClient("http://rpc.cointetris.com/")
	s      = new(stub)
	sessId = ""
	gsHost = ""
)

type tournament struct {
	NumCandidate, CurrNumCandidate int
	AwardGold, AwardSilver         string
	Stat                           string
	Host                           string
	Sponsor                        string
	Tables                         []map[string]interface{}
}

type stub struct {
	CreateSession      func() (string, error)
	SendMailRegister   func(string, string) error
	Register           func(string, string, string, string, string) error
	Login              func(string, string, string) error
	Logout             func(string) error
	GetUserInfo        func(string) (types.User, error)
	UpdateUserPassword func(string, string, string) error
	UpdateUserAvatar   func([]byte, string) error
	Withdraw           func(int, string) (string, error)
	BuyEnergy          func(int, string) error

	Create            func(string, int, string) (int, error)
	Join              func(int, bool, string) (string, error)
	AutoMatch         func(string) (string, string, error)
	GetNormalHall     func(int, int, bool, string) ([]map[string]interface{}, error)
	GetTournamentHall func(int, int, bool, string) (tournament, error)
}

func initRpcClient() {
	client.UseService(&s)
	client.SetKeepAlive(true)
}

func createSession() {
	var err error
	sessId, err = s.CreateSession()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("successfully create session: ", sessId)
}

func register() {
	fmt.Println("please input your email...")
	var email string
	if _, err := fmt.Scanln(&email); err != nil {
		fmt.Println("can not scan email: ", err)
		return
	}
	fmt.Println("sending authen code to ", email)
	if err := s.SendMailRegister(email, sessId); err != nil {
		fmt.Println("can not send email: ", err)
		return
	}
	fmt.Println("please input the authen code...")
	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		fmt.Println("can not scan code: ", err)
		return
	}
	var password, nickname string
	fmt.Println("please input password...")
	if _, err := fmt.Scanln(&password); err != nil {
		fmt.Println("can not scan password: ", err)
		return
	}
	fmt.Println("please input nickname...")
	if _, err := fmt.Scanln(&nickname); err != nil {
		fmt.Println("can not scan nickname: ", err)
		return
	}
	if err := s.Register(email, password, nickname, code, sessId); err != nil {
		fmt.Println("can not register the user: ", err)
		return
	}
	fmt.Printf("register user with email %v and nickname %v, password %v success...\n", email, nickname, password)
}

func login() {
	fmt.Println("please input nickname and password...")
	var nickname, password string
	if _, err := fmt.Scanln(&nickname); err != nil {
		fmt.Println("can not scan nickname: ", err)
		return
	}
	if _, err := fmt.Scanln(&password); err != nil {
		fmt.Println("can not scan password: ", err)
		return
	}
	if err := s.Login(nickname, password, sessId); err != nil {
		fmt.Printf("can not login for nickname %v and password %v: %v\n", nickname, password, err)
		return
	}
	fmt.Println("successfully login...")
}

func logout() {
	if err := s.Logout(sessId); err != nil {
		fmt.Println("logout error: ", err)
		return
	}
	fmt.Println("successfully logout...")
}

func getUserInfo() {
	u, err := s.GetUserInfo(sessId)
	if err != nil {
		fmt.Println("can not get user info: ", err)
		return
	}
	fmt.Printf("successfully get user info:\n%v\n", u)
}

func updateUserAvatar() {
	if err := s.UpdateUserAvatar([]byte("hello"), sessId); err != nil {
		fmt.Println("get error when update avatar:", err)
		return
	}
	fmt.Println("update successful...")
}

func updateUserPassword() {
	var currP, newP string
	fmt.Println("input current password")
	if _, err := fmt.Scanln(&currP); err != nil {
		fmt.Println("can not scan current password:", err)
		return
	}
	fmt.Println("input new password")
	if _, err := fmt.Scanln(&newP); err != nil {
		fmt.Println("can not scan new password:", err)
		return
	}
	if err := s.UpdateUserPassword(currP, newP, sessId); err != nil {
		fmt.Println("can not update user password:", err)
		return
	}
	fmt.Println("update user password successfully")
}

func create() {
	var title string
	var bet int
	fmt.Println("input title...")
	if _, err := fmt.Scanln(&title); err != nil {
		fmt.Println("can not scan title: ", err)
		return
	}
	fmt.Println("input bet...")
	if _, err := fmt.Scanln(&bet); err != nil {
		fmt.Println("can not scan bet: ", err)
		return
	}
	tid, err := s.Create(title, bet, sessId)
	if err != nil {
		fmt.Println("can not create table: ", err)
		return
	}
	fmt.Println("successfully create table: ", tid)
}

func getNormal() {
	normals, err := s.GetNormalHall(10, 1, false, sessId)
	if err != nil {
		fmt.Println("can not get normal hall:", err)
		return
	}
	for _, t := range normals {
		fmt.Printf("%+v\n", t)
		gsHost = t["table_host"].(string)
	}
	fmt.Println("game server host change to: ", gsHost)
}

func getTournament() {
	t, err := s.GetTournamentHall(10, 1, false, sessId)
	if err != nil {
		fmt.Println("can not get tournament hall:", err)
		return
	}
	fmt.Printf("%+v\n", t)
}

// TODO: not finish yet
func join() {
	var tidString, isObString string
	fmt.Println("please input table id...")
	if _, err := fmt.Scanln(&tidString); err != nil {
		fmt.Println("can not scan tidString: ", err)
		return
	}
	fmt.Println("are you ob? true or false")
	if _, err := fmt.Scanln(&isObString); err != nil {
		fmt.Println("can not scan isObString: ", err)
		return
	}
	isOb := isObString == "true"
	tid, err := strconv.Atoi(tidString)
	if err != nil {
		fmt.Println("can not convert to integer table id: ", err)
		return
	}
	token, err := s.Join(tid, isOb, sessId)
	if err != nil {
		fmt.Println("can not join table:", err)
		return
	}
	fmt.Println("get token, connect to socket server...", token)
	handleConn(gsHost, token)
}
