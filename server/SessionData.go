package server

import (
	uuid2 "github.com/gofrs/uuid"
	"github.com/racerxdl/go.fifo"
  "github.com/racerxdl/radioserver/frontends"
  "github.com/racerxdl/radioserver/DSP"
	"github.com/racerxdl/radioserver/protocol"
	"time"
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

	s := &Session{
		IQFifo:      fifo.NewQueue(),
		ID:          ID,
		LastUpdate:  time.Now(),
		CG:          CG,
		fullStopped: false,
  }

  s.frontend = s.ProvisionFrontend(d)
  if s.frontend == nil {
    return nil
  }

	CG.SetOnIQ(func(samples []complex64) {
		if s.IQFifo.Len() < maxFifoBuffs && !s.fullStopped {
			s.IQFifo.Add(samples)
		}
	})

	CG.Start()
  s.frontend.Start()

  return s
}

func (s *Session) ProvisionFrontend(d *protocol.DeviceInfo) frontends.Frontend {
  constructor := frontends.Available[d.DeviceType.String()]
  if constructor == nil {
    return nil
  }

  f := constructor(d)
  f.Init()
  f.SetSamplesAvailableCallback(s.CG.PushSamples)
  return f
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
