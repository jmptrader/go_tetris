package main

import (
	"fmt"
	"math"
	"regexp"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

const (
	sessKeyUserId     = "userId"
	sessKeyRegister   = "register"
	sessKeyForgetPass = "forget"
	sessKeyEmail      = "email"
	minWithdraw       = 1
	minEnergy         = 1
	ratioEnergy2mBTC  = 10
	maxAvatar         = 1 << 18 // 256KB
	defaultEnergy     = 10

	// error const
	errUserNotExist    = "用户 %v 不存在"
	errTableNotExist   = "桌子 %v 不存在, 请加入别的桌子"
	errUnmatchedEmail  = "验证码发到邮箱 %s, 注册邮箱也必须是这个! 请不要这样攻击我们的服务!"
	errExceedMaxAvatar = "头像大小不能超过256KB, 该头像大小为 %dKB."
)

var (
	// regular expressions
	emailReg = regexp.MustCompile(`^[\w.]+\@[\w-]+\.[\w.]+$`)
	passReg  = regexp.MustCompile(`^[\w]{6,22}$`)
	nickReg  = regexp.MustCompile(`^(\p{Han}|\w){2,8}$`)

	// errors
	errEncryptPassword           = fmt.Errorf("密码哈希错误")
	errIncorrectPwd              = fmt.Errorf("密码错误")
	errIncorrectEmailFormat      = fmt.Errorf("邮箱格式错误")
	errIncorrectNicknameFormat   = fmt.Errorf("昵称格式错误, 必须由2~8个中文,英文,数字组成")
	errIncorrectPasswordFormat   = fmt.Errorf("密码格式错误, 必须由6~22个英文,数字组成")
	errEmailExist                = fmt.Errorf("邮箱已经被占用")
	errNicknameExist             = fmt.Errorf("昵称已经被占用")
	errNotLoggedIn               = fmt.Errorf("请先登陆")
	errBalNotSufficient          = fmt.Errorf("余额不足")
	errInvalidBtcAddr            = fmt.Errorf("无效的比特币地址")
	errExceedMinWithdraw         = fmt.Errorf("最低提现额度是 1mBTC")
	errExceedMinEnergy           = fmt.Errorf("最低购买1mBTC能量, 每1mBTC可以充%v能量", ratioEnergy2mBTC)
	errRegisterGetCodeFirst      = fmt.Errorf("需要先发送验证码到邮箱, 才能完成注册")
	errRegisterIncorrectCode     = fmt.Errorf("验证码错误, 请检查邮箱")
	errForgetPassGetCodeFirst    = fmt.Errorf("需要先发送验证码到邮箱, 才能找回密码")
	errForgetPassIncorrectCode   = fmt.Errorf("验证码错误, 请检查邮箱")
	errTableGameIsStarted        = fmt.Errorf("游戏已经开始, 请加入别的桌子")
	errTableIsFull               = fmt.Errorf("桌子已满, 请加入别的桌子进行游戏")
	errUnmatchNumOfFieldAndVal   = fmt.Errorf("两个数组长度不一样")
	errIncorrectType             = fmt.Errorf("更新字段类型只能是字符串, 整形, 二进制流[]byte")
	errAlreadyInGame             = fmt.Errorf("你已经在游戏中, 请先退出再加入另一个游戏")
	errNegativeBet               = fmt.Errorf("赌注不能为负数")
	errCantApplyForNilTournament = fmt.Errorf("暂无争霸赛, 无法加入.")
	errNilTournamentHall         = fmt.Errorf("暂无争霸赛, 无法获得争霸赛桌子信息")
	errCantMatchOpponent         = fmt.Errorf("无匹配对手, 请稍后重试.")
	errNoWorkingGameServer       = fmt.Errorf("当前没有游戏服务器工作")
)

// create session and return session id
func (pubStub) CreateSession() string {
	return session.CreateSession()
}

// send mail, register auth
func (pubStub) SendMailRegister(to string, sessId string) {
	if !emailReg.MatchString(to) {
		panic(errIncorrectEmailFormat)
	}
	if u := getUserByEmail(to); u != nil {
		panic(errEmailExist)
	}
	authenCode := utils.RandString(8)
	session.SetSession(sessKeyRegister, authenCode, sessId)
	session.SetSession(sessKeyEmail, to, sessId)
	f := func() {
		text := "请输入验证码: " + authenCode
		subject := "注册验证"
		if err := utils.SendMail(text, subject, to); err != nil {
			log.Warn("send mail to %v error: %v", to, err)
			log.Warn("text is %v", text)
		}
	}
	pushFunc(f)
}

// send mail, forget password
func (pubStub) SendMailForget(to string, sessId string) {
	if !emailReg.MatchString(to) {
		panic(errIncorrectEmailFormat)
	}
	if u := getUserByEmail(to); u == nil {
		panic(fmt.Errorf(errUserNotExist, to))
	}
	authenCode := utils.RandString(8)
	session.SetSession(sessKeyForgetPass, authenCode, sessId)
	session.SetSession(sessKeyEmail, to, sessId)
	f := func() {
		text := "请输入验证码: " + authenCode
		subject := "找回密码"
		if err := utils.SendMail(text, subject, to); err != nil {
			log.Warn("send mail to %v error: %v", to, err)
			log.Warn("text is %v", text)
		}
	}
	pushFunc(f)
}

// forget password, modify it
func (pubStub) ForgetPassword(newPassword, authenCode string, sessId string) {
	// check authen code
	if code, ok := session.GetSession(sessKeyForgetPass, sessId).(string); !ok {
		panic(errForgetPassGetCodeFirst)
	} else if code != authenCode {
		panic(errForgetPassIncorrectCode)
	}

	if !passReg.MatchString(newPassword) {
		panic(errIncorrectPasswordFormat)
	}
	// get user by email & update it
	email, ok := session.GetSession(sessKeyEmail, sessId).(string)
	if !ok {
		panic(errForgetPassGetCodeFirst)
	}
	u := getUserByEmail(email)
	if u == nil {
		panic(fmt.Errorf(errUserNotExist, email))
	}
	if err := u.Update(types.NewUpdateString(types.UF_Password, utils.Encrypt(newPassword))); err != nil {
		panic(err)
	}
	// delete session
	session.DeleteKey(sessKeyForgetPass, sessId)
	session.DeleteKey(sessKeyEmail, sessId)
	pushFunc(func() { insertOrUpdateUser(u) })
}

// register
func (pubStub) Register(email, password, nickname, authenCode string, sessId string) {
	// check authen code
	if code, ok := session.GetSession(sessKeyRegister, sessId).(string); !ok {
		panic(errRegisterGetCodeFirst)
	} else if code != authenCode {
		panic(errRegisterIncorrectCode)
	}
	// check email
	if tmpEmail, ok := session.GetSession(sessKeyEmail, sessId).(string); !ok {
		panic(errRegisterGetCodeFirst)
	} else if email != tmpEmail {
		panic(fmt.Errorf(errUnmatchedEmail, tmpEmail))
	}

	// check input
	if !emailReg.MatchString(email) {
		panic(errIncorrectEmailFormat)
	}
	if !nickReg.MatchString(nickname) {
		panic(errIncorrectNicknameFormat)
	}
	if !passReg.MatchString(password) {
		panic(errIncorrectPasswordFormat)
	}
	if getUserByEmail(email) != nil {
		panic(errEmailExist)
	}
	if getUserByNickname(nickname) != nil {
		panic(errNicknameExist)
	}
	// generate new bitcoin address for the user
	addr, err := getNewAddress(nickname)
	if err != nil {
		panic(err)
	}
	// encrypt password
	password = utils.Encrypt(password)
	if password == "" {
		panic(errEncryptPassword)
	}
	// delete the authen code
	session.DeleteKey(sessKeyRegister, sessId)
	session.DeleteKey(sessKeyEmail, sessId)
	// add user in cache
	u := types.NewUser(users.GetNextId(), email, password, nickname, addr)
	u.Update(types.NewUpdateInt(types.UF_Energy, defaultEnergy)) // new user get 10 energy
	users.Add(u)
	// async insert into database
	pushFunc(func() { insertOrUpdateUser(u) })
}

// login
func (pubStub) Login(nickname, password string, sessId string) {
	u := getUserByNickname(nickname)
	if u == nil {
		panic(fmt.Errorf(errUserNotExist, nickname))
	}
	if u.Password != utils.Encrypt(password) {
		panic(errIncorrectPwd)
	}
	session.SetSession(sessKeyUserId, u.Uid, sessId)
}

// logout
func (pubStub) Logout(sessId string) {
	// set user to free stat
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		users.SetFree(uid)
	}
	session.DelSession(sessId)
}

// get user info
// update session timestamp
func (pubStub) GetUserInfo(sessId string) *types.User {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		return u
	}
	panic(errNotLoggedIn)
}

// update user password
func (pubStub) UpdateUserPassword(currPass, newPass string, sessId string) {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u.Password != utils.Encrypt(currPass) {
			panic(errIncorrectPwd)
		}
		if !passReg.MatchString(newPass) {
			panic(errIncorrectPasswordFormat)
		}
		if err := u.Update(types.NewUpdateString(types.UF_Password, utils.Encrypt(newPass))); err != nil {
			panic(err)
		}
		pushFunc(func() { insertOrUpdateUser(u) })
		session.SetSession(sessKeyUserId, uid, sessId)
	}
	panic(errNotLoggedIn)
}

// update user avatar
func (pubStub) UpdateUserAvatar(avatar []byte, sessId string) {
	if l := len(avatar); l > maxAvatar {
		panic(fmt.Errorf(errExceedMaxAvatar, l))
	}
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if err := u.Update(types.NewUpdate2dByte(types.UF_Avatar, avatar)); err != nil {
			panic(err)
		}
		pushFunc(func() { insertOrUpdateUser(u) })
		session.SetSession(sessKeyUserId, uid, sessId)
	}
	panic(errNotLoggedIn)
}

// get normal hall
func (pubStub) GetNormalHall(numTableInPage, pageNum int, filterWait bool, sessId string) []map[string]interface{} {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		return normalHall.Wrap(numTableInPage, pageNum, filterWait)
	}
	panic(errNotLoggedIn)
}

// get tournament hall
func (pubStub) GetTournamentHall(numTableInPage, pageNum int, filterWait bool, sessId string) map[string]interface{} {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		return tournamentHall.Wrap(numTableInPage, pageNum, filterWait)
	}
	panic(errNotLoggedIn)
}

// get normal table
func (pubStub) GetNormalTable(tid int, sessId string) map[string]interface{} {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		return normalHall.GetTableById(tid).WrapTable()
	}
	panic(errNotLoggedIn)
}

// get tournament table
func (pubStub) GetTournamentTable(tid int, sessId string) map[string]interface{} {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if tournamentHall == nil {
			panic(errNilTournamentHall)
		}
		return tournamentHall.GetTableById(tid).WrapTable()
	}
	panic(errNotLoggedIn)
}

// join a normal game, play or observe
// actually it is just get a token
func (pubStub) Join(tid int, isOb bool, sessId string) string {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if users.IsBusyUser(uid) {
			panic(errAlreadyInGame)
		}
		t := normalHall.GetTableById(tid)
		if t == nil {
			panic(fmt.Errorf(errTableNotExist, tid))
		}
		if u.GetBalance() < t.GetBet() {
			panic(errBalNotSufficient)
		}
		if u.GetEnergy() <= 0 {
			panic(errInsufficientEnergy)
		}
		if !isOb {
			if t.IsStart() {
				panic(errTableGameIsStarted)
			}
			if t.IsFull() {
				panic(errTableIsFull)
			}
		}
		token, err := utils.GenerateToken(uid, u.Nickname, false, isOb, tid)
		if err != nil {
			panic(err)
		}
		session.SetSession(sessKeyUserId, uid, sessId)
		return token
	}
	panic(errNotLoggedIn)
}

// TODO:
// observe a tournament game
// actually it is just get a token
func (pubStub) ObserveTournament(tid int, sessId string) string {
	if tournamentHall == nil {
		panic(errNilTournamentHall)
	}
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if users.IsBusyUser(uid) {
			panic(errAlreadyInGame)
		}
		t := tournamentHall.GetTableById(tid)
		if t == nil {
			panic(fmt.Errorf(errTableNotExist, tid))
		}
		token, err := utils.GenerateToken(uid, u.Nickname, false, true, tid)
		if err != nil {
			panic(err)
		}
		session.SetSession(sessKeyUserId, uid, sessId)
		return token
	}
	panic(errNotLoggedIn)
}

// automatically match a normal game for user
func (pubStub) AutoMatch(sessId string) (string, string) {
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if users.IsBusyUser(uid) {
			panic(errAlreadyInGame)
		}
		if u.GetEnergy() <= 0 {
			panic(errInsufficientEnergy)
		}
		var table *types.Table = nil
		var gap float64 = 100
		for _, t := range normalHall.Tables.Tables {
			if t.IsFull() {
				continue
			}
			if u.GetBalance() < t.GetBet() {
				continue
			}
			if id := t.Get1pUid(); id != -1 {
				if tGap := math.Abs(float64(getUserById(id).Level - u.Level)); tGap < gap {
					gap = tGap
					table = t
					if tGap == 0 {
						break
					}
				}
				continue
			}
			if id := t.Get2pUid(); id != -1 {
				if tGap := math.Abs(float64(getUserById(id).Level - u.Level)); tGap < gap {
					gap = tGap
					table = t
					if tGap == 0 {
						break
					}
				}
				continue
			}
		}
		if table == nil {
			panic(errCantMatchOpponent)
		}
		token, err := utils.GenerateToken(uid, u.Nickname, false, false, table.TId)
		if err != nil {
			panic(err)
		}
		session.SetSession(sessKeyUserId, uid, sessId)
		return table.GetHost(), token
	}
	panic(errNotLoggedIn)
}

// create a game
func (pubStub) Create(title string, bet int, sessId string) int {
	if bet < 0 {
		panic(errNegativeBet)
	}
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if users.IsBusyUser(uid) {
			panic(errAlreadyInGame)
		}
		if u.GetBalance() < bet {
			panic(errBalNotSufficient)
		}
		if u.GetEnergy() <= 0 {
			panic(errInsufficientEnergy)
		}
		id := normalHall.NextTableId()
		ip := clients.BestServer()
		if ip == "" {
			panic(errNoWorkingGameServer)
		}
		host := ip + ":" + gameServerSocketPort
		if err := clients.GetStub(ip).Create(id); err != nil {
			panic(err)
		}
		if err := normalHall.NewTable(id, title, host, bet); err != nil {
			panic(err)
		}
		return id
	}
	panic(errNotLoggedIn)
}

// TODO:
// apply for a tournament
func (pubStub) Apply(sessId string) (string, string) {
	if tournamentHall == nil {
		panic(errCantApplyForNilTournament)
	}
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		if users.IsBusyUser(uid) {
			panic(errAlreadyInGame)
		}
		token, err := utils.GenerateToken(uid, u.Nickname, true, false, -1)
		if err != nil {
			panic(err)
		}
		session.SetSession(sessKeyUserId, uid, sessId)
		return tournamentHall.GetHost(), token
	}
	panic(errNotLoggedIn)
}

// withdraw
// amount in mBTC
func (pubStub) Withdraw(amount int, address string, sessId string) string {
	// check if amount >= minWithdraw
	if amount < minWithdraw {
		panic(errExceedMinWithdraw)
	}
	// check btc address
	if isValid, err := validateAddress(address); err != nil {
		panic(err)
	} else if !isValid {
		panic(errInvalidBtcAddr)
	}
	// check if user logged in
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		// check balance
		if u.GetBalance() < amount {
			panic(errBalNotSufficient)
		}
		// send coin
		txid, err := sendBitcoin(address, amount)
		if err != nil {
			panic(err)
		}
		// update user in the cache
		if err := u.Update(types.NewUpdateInt(types.UF_Balance, u.GetBalance()-amount)); err != nil {
			panic(err)
		}
		pushFunc(func() { insertWithdraw(txid, u.Nickname, address, amount) })
		session.SetSession(sessKeyUserId, uid, sessId)
		return txid
	}
	panic(errNotLoggedIn)
}

// buy energy
func (pubStub) BuyEnergy(amountOfmBTC int, sessId string) {
	// check if amountOfmBTC >= minEnergy
	if amountOfmBTC < minEnergy {
		panic(errExceedMinEnergy)
	}
	// check if user logged in
	if uid, ok := session.GetSession(sessKeyUserId, sessId).(int); ok {
		u := getUserById(uid)
		if u == nil {
			panic(fmt.Errorf(errUserNotExist, uid))
		}
		// check balance
		if u.GetBalance() < amountOfmBTC {
			panic(errBalNotSufficient)
		}
		// update user in the cache
		if err := u.Update(types.NewUpdateInt(types.UF_Balance, u.GetBalance()-amountOfmBTC),
			types.NewUpdateInt(types.UF_Energy, u.GetEnergy()+amountOfmBTC*ratioEnergy2mBTC)); err != nil {
			panic(err)
		}
		pushFunc(func() { buyEnergy(u.Uid, amountOfmBTC) })
		session.SetSession(sessKeyUserId, uid, sessId)
	}
	panic(errNotLoggedIn)
}

// online players
func (pubStub) NumOfOnlinePlayer() int {
	// connections hold by all game servers
	// return clients.TotalConnections()

	// judge by session
	return session.NumOfOnlineUsers()
}

// 不需要sessionId 的函数
var notNeedSessFunc = map[string]bool{
	"NumOfOnlinePlayer": true,
	"CreateSession":     true,
}
