package types

import (
	"encoding/json"
	"sync"
)

// observers
type obs struct {
	users map[int]*User
	mu    sync.RWMutex
}

func NewObs() *obs {
	return &obs{
		users: make(map[int]*User),
	}
}

func (this *obs) Wrap() []*User {
	this.mu.RLock()
	defer this.mu.RUnlock()
	us := make([]*User, 0)
	for _, u := range this.users {
		us = append(us, u)
	}
	return us
}

// join a new observer
func (this *obs) Join(u *User) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.users[u.Uid] = u
}

// quit an observer
func (this *obs) Quit(uId int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	// if u := this.users[uId]; u != nil {
	// 	u.Close()
	// }
	delete(this.users, uId)
}

// quit all users
func (this *obs) QuitAll() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for uid, _ := range this.users {
		// if u != nil {
		// 	u.Close()
		// }
		delete(this.users, uid)
	}
}

// get user by id
func (this *obs) GetUserById(uid int) *User {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.users[uid]
}

// get all observers uid
func (this *obs) GetAll() []int {
	us := make([]int, 0)
	for i, _ := range this.users {
		us = append(us, i)
	}
	return us
}

// check if a user is in obs
func (this *obs) IsUserExist(uid int) bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.users[uid] != nil
}

// get all observers' connection
func (this *obs) GetConns() []*User {
	this.mu.RLock()
	defer this.mu.RUnlock()
	conns := make([]*User, 0)
	for _, u := range this.users {
		conns = append(conns, u)
	}
	return conns
}

func (this *obs) MarshalJSON() ([]byte, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	res := make([]*User, 0)
	for _, v := range this.users {
		res = append(res, v)
	}
	return json.Marshal(res)
}
