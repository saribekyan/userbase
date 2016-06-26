package auth

import (
	"sync"
	"time"
)

type SessionManager struct {
	sessions map[string]Session

	mu sync.Mutex

	killCh chan struct{}
}

type Session struct {
	username string
	key      string
	lastSeen time.Time
	created  time.Time

	// channel that is signaled when the session is loaded again
	aliveCh chan struct{}
}

func (sm *SessionManager) NewSession(username string) string {
	session := Session{
		username: username,
		key:      randomString(SESSION_KEY_LEN),
		lastSeen: time.Now(),
		created:  time.Now(),
		aliveCh:  make(chan struct{})}

	sm.mu.Lock()
	sm.sessions[session.key] = session
	sm.mu.Unlock()

	// keep the session alive for as long as needed
	go func(session Session) {
		for {
			select {
			case <-session.aliveCh:
				// pass
			case <-time.After(SESSION_TIMEOUT):
				sm.RemoveSession(session.key)
				return

			case <-sm.killCh:
				return
			}
		}
	}(session)

	return session.key
}

func (sm *SessionManager) AuthenticateSessionKey(sessionKey string) (string, bool) {
	sm.mu.Lock()
	session, isIn := sm.sessions[sessionKey]
	sm.mu.Unlock()
	if isIn {
		session.lastSeen = time.Now()
		session.aliveCh <- struct{}{}
		return session.username, true
	}
	return "", false
}

func (sm *SessionManager) RemoveSession(key string) {
	sm.mu.Lock()
	delete(sm.sessions, key)
	sm.mu.Unlock()
}

func (sm *SessionManager) NSessionsOfUser(username string) int {
	n := 0
	sm.mu.Lock()
	for _, session := range sm.sessions {
		if session.username == username {
			n += 1
		}
	}
	sm.mu.Unlock()
	return n
}

func (sm *SessionManager) Kill() {
	close(sm.killCh)
}

func MakeSM() *SessionManager {
	sm := &SessionManager{}
	sm.sessions = make(map[string]Session)
	sm.killCh = make(chan struct{})
	return sm
}
