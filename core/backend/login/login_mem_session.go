package login

import (
	"sync"
	"time"
)

var _msm = &memSessionManager{
	tsm: new(sync.Map),
	stm: new(sync.Map),
	sem: new(sync.Map),
}

func SMInstance() SessionManager { return _msm }

type SessionManager interface {
	Store(string, string)
	Load(string) (string, bool)
	Keep(string)
}

type memSessionManager struct {
	tsm *sync.Map // token - session map
	stm *sync.Map // session - token map
	sem *sync.Map // session - expire time map
}

func (msm *memSessionManager) Store(token, session string) {
	msm.tsm.Store(token, session)
	msm.stm.Store(session, token)
	msm.sem.Store(session, time.Now().Add(time.Minute*time.Duration(5)).Unix())
}

func (msm *memSessionManager) Load(session string) (string, bool) {
	v, e := msm.stm.Load(session)
	if !e {
		return "", false
	}

	et, ee := msm.sem.Load(session)
	if !ee {
		return "", true
	}

	return v.(string), time.Now().After(time.Unix(et.(int64), 0))
}

func (msm *memSessionManager) Keep(session string) {
	if _, exist := msm.stm.Load(session); !exist {
		return
	}

	msm.sem.Store(session, time.Now().Add(time.Minute*time.Duration(5)).Unix())
}
