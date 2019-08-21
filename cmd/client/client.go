package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/radioserver/protocol"
	"google.golang.org/grpc"
)

var log = slog.Scope("RadioClient")

var empty = &protocol.Empty{}

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

  // Get Server Meta

	server, err := client.ServerInfo(ctx, empty)
	if err != nil {
		log.Fatal(err)
	}

  serverInfo, _ := json.MarshalIndent(server, "", "   ")
	log.Info("Server Info: %s", serverInfo)

  // Get Devices List
	dls, err := client.List(ctx, empty)
	if err != nil {
		log.Fatal(err)
	}

  deviceList, _ := json.MarshalIndent(dls, "", "   ")
	log.Info("Server Info: %s", deviceList)
}
