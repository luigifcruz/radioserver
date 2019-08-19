package server

import (
	uuid2 "github.com/gofrs/uuid"
	"github.com/racerxdl/go.fifo"
  "github.com/racerxdl/radioserver/frontends"
  "github.com/racerxdl/radioserver/DSP"
	"github.com/racerxdl/radioserver/protocol"
	"time"
  "fmt"
)

const (
	expirationTime = time.Second * 120
	maxFifoBuffs   = 4096
)

type Session struct {
  ID         string
	LastUpdate time.Time

  frontend    frontends.Frontend

	IQFifo      *fifo.Queue
	CG          *DSP.ChannelGenerator

	fullStopped bool
}

func GenerateSession(d *protocol.DeviceInfo) *Session {
	u, _ := uuid2.NewV4()
	ID := u.String()

	CG := DSP.CreateChannelGenerator()

  log.Info("Initializing Frontend")
  frontend := frontends.CreateLimeSDRFrontend(0)
  frontend.Init()

  frontend.SetCenterFrequency(96900000)
  frontend.SetSampleRate(3000000)
  frontend.SetGain(60)
  frontend.Start()

	s := &Session{
		IQFifo:      fifo.NewQueue(),
		ID:          ID,
		LastUpdate:  time.Now(),
		CG:          CG,
		fullStopped: false,
	  frontend:    frontend,
  }

	CG.SetOnIQ(func(samples []complex64) {
		if s.IQFifo.Len() < maxFifoBuffs && !s.fullStopped {
			s.IQFifo.Add(samples)
		}
	})
	CG.Start()

  frontend.SetSamplesAvailableCallback(s.CG.PushSamples)

  return s
}

func (s *Session) Expired() bool {
	return time.Since(s.LastUpdate) > expirationTime
}

func (s *Session) KeepAlive() {
	s.LastUpdate = time.Now()
}

func (s *Session) IsFullStopped() bool {
	return s.fullStopped
}

func (s *Session) FullStop() {
  s.frontend.Stop()
  s.CG.StopIQ()
	s.CG.Stop()
	s.fullStopped = true
}
