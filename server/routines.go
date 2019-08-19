package server

import "time"

const (
	routinesInterval     = time.Second * 2
	sessionCheckInterval = time.Second * 10
)

func (rs *RadioServer) routines() {
	log.Info("RadioServer Routines Started")
	rs.lastSessionChecks = time.Now()
	for rs.running {
		rs.checkSessions()
		time.Sleep(routinesInterval)
	}
	log.Warn("RadioServer Routines Stopped")
}

func (rs *RadioServer) checkSessions() {
	if time.Since(rs.lastSessionChecks) < sessionCheckInterval {
		return
	}

	rs.sessionLock.Lock()
	defer rs.sessionLock.Unlock()

	for token, session := range rs.sessions {
		if session.Expired() {
			delete(rs.sessions, token)
			go session.FullStop()
		}
	}
	rs.lastSessionChecks = time.Now()
}
