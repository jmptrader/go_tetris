package types

import (
	"encoding/json"
	"fmt"

	"reflect"
	"sync"
	"time"
)

var levels = map[int]string{}
var errUserNotExist = "User %v does not exist"

// Users is a cache which has an expire(in days)
type Users struct {
	// index by userId, email
	users     map[int]*User
	emails    map[string]*User
	nicknames map[string]*User
	mu        sync.RWMutex

	// user who is playing or observing a game
	busyUsers map[int]int64
	bmu       sync.RWMutex

	// generate next user Id
	nextId int
	nmu    sync.Mutex

	// expire days
	expire int
}

func NewUsers() *Users {
	us := &Users{
		users:     make(map[int]*User),
		emails:    make(map[string]*User),
		nicknames: make(map[string]*User),
		busyUsers: make(map[int]int64),
	}
	return us.init()
}

// init users cache
// no expire
func (us *Users) init() *Users {
	// go us.releaseExpire()
	return us
}

func (us *Users) releaseExpire() {
	for {
		us.del(func() []int {
			us.mu.RLock()
			defer us.mu.RUnlock()
			ids := make([]int, 0)
			for uid, u := range us.users {
				if (int(time.Now().Unix()) - u.Updated) > us.expire*3600*24 {
					ids = append(ids, uid)
				}
			}
			return ids
		}()...)
		time.Sleep(time.Hour)
	}
}

// set next id
func (us *Users) SetNextId(val int) {
	us.nmu.Lock()
	defer us.nmu.Unlock()
	us.nextId = val
}

// get next id
func (us *Users) GetNextId() int {
	us.nmu.Lock()
	defer us.nmu.Unlock()
	return us.nextId
}

var errAlreadyExist = "User %v is already exist"

// add new users into the cache
func (us *Users) Add(users ...*User) error {
	us.mu.Lock()
	defer us.mu.Unlock()
	// first check if the uid, email is already exist
	// if exist, return errAlreadyExist
	// else insert and return nil
	for _, u := range users {
		if us.users[u.Uid] != nil {
			return fmt.Errorf(errAlreadyExist, u.Uid)
		}
		if us.emails[u.Email] != nil {
			return fmt.Errorf(errAlreadyExist, u.Email)
		}
		if us.nicknames[u.Nickname] != nil {
			return fmt.Errorf(errAlreadyExist, u.Nickname)
		}
		us.users[u.Uid] = u
		us.emails[u.Email] = u
		us.nicknames[u.Nickname] = u
		// set next id
		if us.GetNextId() <= u.Uid {
			us.SetNextId(u.Uid + 1)
		}
	}
	return nil
}

// if email exist
func (us *Users) IsEmailExist(email string) bool {
	return us.GetByEmail(email) != nil
}

// is nickname exist
func (us *Users) IsNicknameExist(nickname string) bool {
	return us.GetByNickname(nickname) != nil
}

// delete users from cache
func (us *Users) del(uid ...int) {
	us.mu.Lock()
	defer us.mu.Unlock()
	for _, v := range uid {
		u, ok := us.users[v]
		if ok {
			delete(us.emails, u.Email)
			delete(us.nicknames, u.Nickname)
			delete(us.users, v)
		}
	}
}

// get all users
func (us *Users) GetAllUsers() []*User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	users := make([]*User, 0)
	for _, u := range us.users {
		users = append(users, u)
	}
	return users
}

// get a user
func (us *Users) GetById(uid int) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.users[uid]
}

func (us *Users) GetByEmail(email string) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.emails[email]
}

func (us *Users) GetByNickname(nickname string) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.nicknames[nickname]
}

// set users in busy mode
func (us *Users) SetBusy(uids ...int) {
	us.bmu.Lock()
	defer us.bmu.Unlock()
	t := time.Now().Unix()
	for _, uid := range uids {
		us.busyUsers[uid] = t
	}
}

// set users in free mode
func (us *Users) SetFree(uids ...int) {
	us.bmu.Lock()
	defer us.bmu.Unlock()
	for _, uid := range uids {
		delete(us.busyUsers, uid)
	}
}

// check if user is in busy mode
func (us *Users) IsBusyUser(uid int) bool {
	us.bmu.RLock()
	defer us.bmu.RUnlock()
	return us.busyUsers[uid] != 0
}

// interface for updating user information
type UpdateInterface interface {
	Field() string
	Val() interface{}
}

// update string field
type updateString struct {
	field string
	val   string
}

func NewUpdateString(field string, val string) updateString {
	return updateString{
		field: field,
		val:   val,
	}
}

func (us updateString) Field() string {
	return us.field
}

func (us updateString) Val() interface{} {
	return us.val
}

// update int field
type updateInt struct {
	field string
	val   int
}

func NewUpdateInt(field string, val int) updateInt {
	return updateInt{
		field: field,
		val:   val,
	}
}

func (ui updateInt) Field() string {
	return ui.field
}

func (ui updateInt) Val() interface{} {
	return ui.val
}

// update []byte field
type update2dByte struct {
	field string
	val   []byte
}

func NewUpdate2dByte(field string, val []byte) update2dByte {
	return update2dByte{
		field: field,
		val:   val,
	}
}

func (u2b update2dByte) Field() string {
	return u2b.field
}

func (u2b update2dByte) Val() interface{} {
	return u2b.val
}

// update
// if not update statement specified, then only update its Updated
func (us *Users) Update(uid int, upts ...UpdateInterface) error {
	us.mu.Lock()
	defer us.mu.Unlock()
	u, ok := us.users[uid]
	if !ok {
		return fmt.Errorf(errUserNotExist, uid)
	}
	return u.Update(upts...)
}

// give energy to all users
func (us *Users) EnergyGiveout(energy int) {
	for uid, u := range us.users {
		if u.GetEnergy() >= energy {
			continue
		}
		us.Update(uid, NewUpdateInt(UF_Energy, energy))
	}
}

// user fields which could be updated
const (
	errCantUpdateField = "Can not update the user field %v"
	UF_Nickname        = "Nickname"
	UF_Avatar          = "Avatar"
	UF_Password        = "Password"
	UF_Energy          = "Energy"
	UF_Level           = "Level"
	UF_Win             = "Win"
	UF_Lose            = "Lose"
	UF_Addr            = "Addr"
	UF_Balance         = "Balance"
	UF_Freezed         = "Freezed"
	UF_Updated         = "Updated"
)

// initialize user field cache
func init() {
	u := User{}
	typ := reflect.TypeOf(u)
	val := reflect.ValueOf(u)
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() && typ.Field(i).Tag.Get("fixed") != "true" {
			userFields[typ.Field(i).Name] = true
		}
	}
}

var userFields = make(map[string]bool)

func canUserFieldUpdate(field string) bool {
	return userFields[field]
}

// fixed tag means it is not going to update the field by reflect method
type User struct {
	Uid      int `fixed:"true"`
	Avatar   []byte
	Email    string `fixed:"true"`
	Password string
	Nickname string `fixed:"true"`
	Energy   int
	Level    int
	Win      int
	Lose     int
	Addr     string `fixed:"true"`
	Balance  int
	Freezed  int
	Updated  int
	// readConn, writeConn *net.TCPConn
	mu sync.Mutex
}

func (u User) Wrap() map[string]interface{} {
	u.mu.Lock()
	defer u.mu.Unlock()
	length, exp := 1, 0
	if u.Level != 0 {
		// length = 2*u.Level - 1
		length = 2 * u.Level
	}
	if u.Level > 1 {
		exp = u.Win - (u.Level-1)*(u.Level-1)
	}
	return map[string]interface{}{
		"uid":      u.Uid,
		"avatar":   u.Avatar,
		"email":    u.Email,
		"nickname": u.Nickname,
		"energy":   u.Energy,
		"level":    u.Level,
		"win":      u.Win,
		"lose":     u.Lose,
		"addr":     u.Addr,
		"balance":  u.Balance,
		"length":   length,
		"exp":      exp,
	}
}

func (u User) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

func NewUser(uid int, email, password, nickname, addr string) *User {
	return &User{
		Uid:      uid,
		Email:    email,
		Password: password,
		Nickname: nickname,
		Addr:     addr,
		Updated:  int(time.Now().Unix()),
	}
}

var ErrNilReadConn = fmt.Errorf("the read tcp connection is nil")
var ErrNilWriteConn = fmt.Errorf("the write tcp connection is nil")

// get uid
func (u *User) GetUid() int {
	if u == nil {
		return -1
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Uid
}

// get current energy
func (u *User) GetEnergy() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Energy
}

// get current balance
func (u *User) GetBalance() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Balance
}

// get current freezed
func (u *User) GetFreezed() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Freezed
}

// update user
func (u *User) Update(upts ...UpdateInterface) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	// check if field really exist
	for _, v := range upts {
		if !canUserFieldUpdate(v.Field()) {
			return fmt.Errorf(errCantUpdateField, v.Field())
		}
	}
	// update it
	u.Updated = int(time.Now().Unix())
	for _, v := range upts {
		reflect.Indirect(reflect.ValueOf(u)).FieldByName(v.Field()).Set(reflect.ValueOf(v.Val()))
	}
	return nil
}

// // set read tcp connection
// func (this *User) SetReadConn(conn *net.TCPConn) {
// 	if this == nil {
// 		log.Println("the user is nil, can not set read connection")
// 		return
// 	}
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	this.readConn = conn
// }
//
// // set write tcp connection
// func (this *User) SetWriteConn(conn *net.TCPConn) {
// 	if this == nil {
// 		log.Println("the user is nil, can not set write connection")
// 		return
// 	}
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	this.writeConn = conn
// }
//
// // get read tcp connection
// func (this *User) GetReadConn() *net.TCPConn {
// 	if this == nil {
// 		log.Println("the user is nil, can not get read connection")
// 		return nil
// 	}
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	return this.readConn
// }
//
// // get write tcp connection
// func (this *User) GetWriteConn() *net.TCPConn {
// 	if this == nil {
// 		log.Println("the user is nil, can not get write connection")
// 		return nil
// 	}
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	return this.writeConn
// }
//
// // set read timeout in secs
// func (this *User) SetReadTimeoutInSecs(n int) error {
// 	if c := this.GetReadConn(); c != nil {
// 		return c.SetReadDeadline(time.Now().Add(time.Duration(n) * time.Second))
// 	}
// 	return ErrNilReadConn
// }
//
// // set write timeout in secs
// func (this *User) SetWriteTimeoutInSecs(n int) error {
// 	if c := this.GetWriteConn(); c != nil {
// 		return c.SetWriteDeadline(time.Now().Add(time.Duration(n) * time.Second))
// 	}
// 	return ErrNilWriteConn
// }
//
// // close
// func (this *User) Close() error {
// 	var err1, err2 error
// 	if c := this.GetReadConn(); c != nil {
// 		err1 = c.Close()
// 	} else {
// 		err1 = ErrNilReadConn
// 	}
// 	if c := this.GetWriteConn(); c != nil {
// 		err2 = c.Close()
// 	} else {
// 		err2 = ErrNilWriteConn
// 	}
// 	if err1 != nil || err2 != nil {
// 		return fmt.Errorf("%v\n%v", err1, err2)
// 	}
// 	return nil
// }

// marshal the user
func (this User) MarshalJSON() ([]byte, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	return json.Marshal(map[string]interface{}{
		"id":       this.Uid,
		"email":    this.Email,
		"nickname": this.Nickname,
		"level":    this.Level,
		"win":      this.Win,
		"lose":     this.Lose,
		"addr":     this.Addr,
		"balance":  this.Balance,
		"updated":  this.Updated,
	})
}

// sql generator
// update or insert
func (this *User) SqlGeneratorUpdate() (sql string, args []interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	typ := reflect.TypeOf(this).Elem()
	val := reflect.Indirect(reflect.ValueOf(this))
	updates := make([]string, 0)
	args = make([]interface{}, 0)
	sql = "INSERT INTO users("
	var l int
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() {
			if i != 0 {
				sql += ", "
			}

			l++
			f := typ.Field(i).Name
			if canUserFieldUpdate(f) {
				updates = append(updates, f)
			}
			sql += f
			args = append(args, val.Field(i).Interface())
		}
	}
	sql += ") VALUES("
	for i := 0; i < l; i++ {
		if i != 0 {
			sql += ", "
		}
		sql += "?"
	}
	sql += ") ON DUPLICATE KEY UPDATE "
	for i, v := range updates {
		if i != 0 {
			sql += ", "
		}
		sql += v + " = ?"
		args = append(args, val.FieldByName(v).Interface())
	}
	return
}

// create table
func (this User) SqlGeneratorCreate() (sqls []string) {
	sql := "CREATE TABLE users (\n"
	typ := reflect.TypeOf(this)
	val := reflect.ValueOf(this)
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() {
			t := typ.Field(i)
			sql += t.Name + " "
			switch t.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				sql += "INT, "
			case reflect.String:
				sql += "VARCHAR(255), "
			case reflect.Slice:
				sql += "BLOB, "
			default:
				panic("can not generate sql to create table")
			}
			sql += "\n"
		}
	}
	sql += "PRIMARY KEY (Uid)\n"
	sql += ") ENGINE=innoDB;\n"
	sql += "\n"
	sqls = make([]string, 0)
	sqls = append(sqls, sql)
	sqls = append(sqls, "CREATE UNIQUE INDEX uni_index_email ON users (Email);")
	sqls = append(sqls, "CREATE UNIQUE INDEX uni_index_nickname ON users (Nickname);")
	return
}
