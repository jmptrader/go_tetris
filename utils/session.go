package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	cookieSessId = "sessId"
)

// session store
type sessionStore struct {
	sess           map[string]*session // sessionId -> *session
	expireInMinute int64               // minimum 30 minutes
	mu             sync.RWMutex
}

func NewSessionStore() *sessionStore {
	ss := &sessionStore{
		sess:           make(map[string]*session),
		expireInMinute: 120,
	}
	return ss.init()
}

// session store initialization
func (ss *sessionStore) Init(sess map[string]map[string]interface{}) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	for sessId, ses := range sess {
		ss.sess[sessId] = newSession()
		for key, val := range ses {
			ss.sess[sessId].set(key, val)
		}
	}
}

// session store start
func (ss *sessionStore) init() *sessionStore {
	go ss.gc()
	return ss
}

// delete expire sessions
func (ss *sessionStore) gc() {
	getExpire := func() []string {
		ss.mu.RLock()
		defer ss.mu.RUnlock()
		tNow := time.Now().Unix()
		sss := make([]string, 0)
		for sessId, v := range ss.sess {
			if (tNow-v.updated)/60 > ss.expireInMinute {
				sss = append(sss, sessId)
			}
		}
		return sss
	}
	for {
		time.Sleep(1 * time.Minute)
		ss.delSession(getExpire()...)
	}
}

// online users
func (ss *sessionStore) NumOfOnlineUsers() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.sess)
}

// generate unique session id
func (ss *sessionStore) generateSessionId() string {
	return RandString(32)
}

// check if session id is already exist
func (ss *sessionStore) IsSessIdExist(sessId string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sess[sessId] != nil
}

// delete session from session store
func (ss *sessionStore) delSession(sessionIds ...string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	for _, sessionId := range sessionIds {
		delete(ss.sess, sessionId)
	}
}

// create a session and return session id
func (ss *sessionStore) CreateSession() string {
	var sessId = ""
	for sessId == "" || ss.IsSessIdExist(sessId) {
		sessId = ss.generateSessionId()
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sess[sessId] = newSession()
	return sessId
}

// get all session to store in db
func (ss *sessionStore) GetAllSession() map[string]map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	res := make(map[string]map[string]interface{})
	for sessId, sess := range ss.sess {
		res[sessId] = make(map[string]interface{})
		for key, val := range sess.vals {
			res[sessId][key] = val
		}
	}
	return res
}

// store data in session
// update updated
func (ss *sessionStore) SetSession(key string, val interface{}, sessId string) {
	if !ss.IsSessIdExist(sessId) {
		return
	}
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	sess := ss.sess[sessId]
	sess.set(key, val)
}

// delete data from session
// update updated
func (ss *sessionStore) DeleteKey(key string, sessId string) {
	if !ss.IsSessIdExist(sessId) {
		return
	}
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	sess := ss.sess[sessId]
	sess.del(key)
}

// del the session id
func (ss *sessionStore) DelSession(sessId string) {
	ss.delSession(sessId)
}

// get data from session
func (ss *sessionStore) GetSession(key string, sessId string) interface{} {
	if !ss.IsSessIdExist(sessId) {
		return nil
	}
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sess[sessId].get(key)
}

// String for testing
func (ss *sessionStore) String() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	str := ""
	for sessId, sess := range ss.sess {
		str += sessId + " --> \n"
		for name, val := range sess.vals {
			str += fmt.Sprintf("\t%v -> %v\n", name, val)
		}
	}
	return str
}

// session
type session struct {
	updated int64
	vals    map[string]interface{}
	mu      sync.RWMutex
}

func newSession() *session {
	return &session{
		updated: time.Now().Unix(),
		vals:    make(map[string]interface{}),
	}
}

func (s *session) set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals[key] = val
	s.updated = time.Now().Unix()
}

func (s *session) get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vals[key]
}

func (s *session) del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vals, key)
	s.updated = time.Now().Unix()
}
