package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/radioserver/protocol"
	"google.golang.org/grpc"
	"time"
)

var log = slog.Scope("RadioClient")

var empty = &protocol.Empty{}
var loginInfo *protocol.LoginData

func GetSamples(client protocol.RadioServerClient, stop chan bool) {
	ctx := context.Background()
	iqClient, err := client.SmartIQ(ctx, &protocol.ChannelConfig{
		LoginInfo:       loginInfo,
		CenterFrequency: 106.3e6,
		DecimationStage: 1,
	})

	if err != nil {
		log.Fatal(err)
	}

	for {
		data, err := iqClient.Recv()
		if err != nil {
			log.Error(err)
			break
		}
		log.Info("Received %d samples!", len(data.Samples))
	}

	stop <- true
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:4050", opts...)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	client := protocol.NewRadioServerClient(conn)

	ctx := context.Background()

	ret, err := client.Hello(ctx, &protocol.DeviceInfo{})

	if err != nil {
		log.Fatal(err)
	}

	loginInfo = ret.Login

	log.Info("Status: %s", ret.Status)
	log.Info("Token: %s", ret.Login.Token)

	server, err := client.ServerInfo(ctx, empty)

	if err != nil {
		log.Fatal(err)
	}

	serverInfo, _ := json.MarshalIndent(server, "", "   ")

	log.Info("Server Info: %s", serverInfo)

	stop := make(chan bool, 1)

	go GetSamples(client, stop)

	stopTimer := time.NewTimer(time.Second * 60)
	running := true

	for running {
		select {
		case <-stopTimer.C:
			running = false
		case <-stop:
			running = false
		}
	}

	r, err := client.Bye(ctx, loginInfo)

	if err != nil {
		log.Fatal(err)
	}

	log.Info(r.Message)
}
