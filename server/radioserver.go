package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/luigifreitas/radioserver"
	"github.com/luigifreitas/radioserver/protocol"
	"github.com/quan-to/slog"
	"google.golang.org/grpc"
)

var log = slog.Scope("RadioServer")

type RadioServer struct {
	serverInfo *protocol.ServerInfoData

	sessions    map[string]*Session
	sessionLock sync.Mutex
	grpcServer  *grpc.Server

	running           bool
	lastSessionChecks time.Time
}

func MakeRadioServer(serverName string) *RadioServer {
	rs := &RadioServer{
		serverInfo: &protocol.ServerInfoData{
			Name: serverName,
			Version: &protocol.Version{
				Major: uint32(radioserver.ServerVersion.Major),
				Minor: uint32(radioserver.ServerVersion.Minor),
				Hash:  radioserver.ServerVersion.Hash,
			},
		},
		sessions:    map[string]*Session{},
		sessionLock: sync.Mutex{},
	}

	return rs
}

func (rs *RadioServer) Listen(address string) error {
	if rs.grpcServer != nil {
		return fmt.Errorf("server already runing")
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	rs.grpcServer = grpc.NewServer()

	protocol.RegisterRadioServerServer(rs.grpcServer, rs)
	rs.running = true
	go rs.routines()
	go rs.serve(lis)
	return nil
}

func (rs *RadioServer) serve(conn net.Listener) {
	err := rs.grpcServer.Serve(conn)
	if err != nil {
		log.Error("RPC Error: %s", err)
	}
	rs.Stop()
}

func (rs *RadioServer) Stop() {
	if rs.grpcServer == nil {
		return
	}
	log.Info("Stopping RPC Server")
	rs.grpcServer.Stop()
	rs.grpcServer = nil
	rs.running = false
}
