package server

import (
  "context"
  "fmt"
	"github.com/racerxdl/radioserver/protocol"
	"runtime"
	"sync"
	"time"
)

// region GRPC Stuff

func (rs *RadioServer) List(ctx context.Context, s *protocol.Empty) (*protocol.DeviceList, error) {
  return nil, nil
}

func (rs *RadioServer) Provision(ctx context.Context, d *protocol.DeviceInfo) (*protocol.DeviceInfo, error) {
	rs.sessionLock.Lock()
	defer rs.sessionLock.Unlock()

	s := GenerateSession(d)
  if s == nil {
    return nil, fmt.Errorf("error provisioning")
  }

  rs.sessions[s.ID] = s
	log.Info("Provisioned %s!", s.ID)
  d.Session = s.ID

	return d, nil
}

func (rs *RadioServer) Destroy(ctx context.Context, sid *protocol.Session) (*protocol.Empty, error) {
	rs.sessionLock.Lock()
	defer rs.sessionLock.Unlock()

	s := rs.sessions[sid.Token]
	if s == nil {
		return nil, fmt.Errorf("not logged in")
	}

	delete(rs.sessions, sid.Token)
	s.FullStop()

	log.Info("Destroyed %s!", s.ID)
	return nil, nil
}

func (rs *RadioServer) ServerInfo(context.Context, *protocol.Empty) (*protocol.ServerInfoData, error) {
	return rs.serverInfo, nil
}

func (rs *RadioServer) Tune(ctx context.Context, cc *protocol.StreamConfig) (*protocol.StreamConfig, error) {
  return cc, nil
}

func (rs *RadioServer) RXIQ(sid *protocol.Session, server protocol.RadioServer_RXIQServer) error {
	s := rs.sessions[sid.Token]
	if s.CG.IQRunning() {
		return fmt.Errorf("already running")
	}

	s.CG.StartIQ()
	delete(rs.sessions, sid.Token)
  defer s.FullStop()

	lastNumSamples := 0
	pool := sync.Pool{
		New: func() interface{} {
			return make([]float32, lastNumSamples)
		},
	}

	for {
		for s.IQFifo.Len() > 0 {
			samples := s.IQFifo.Next().([]complex64)
			pb := protocol.MakeIQDataWithPool(samples, pool)
			if err := server.Send(pb); err != nil {
				log.Error("Error sending samples to %s: %s", s.ID, err)
				return err
			}
			s.KeepAlive()

			if len(pb.Samples) != lastNumSamples {
				lastNumSamples = len(pb.Samples)
			}

			pool.Put(pb.Samples) // If the size is not correct, MakeIQDataWithPool will discard or trim it

			if s.IsFullStopped() {
				log.Error("Session Expired")
				return fmt.Errorf("session expired")
			}
			runtime.Gosched()
		}
		time.Sleep(time.Millisecond)
	}
}

// endregion
