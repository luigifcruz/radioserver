package main

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/luigifreitas/radioserver/protocol"
	"github.com/quan-to/slog"
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
	log.Info("Available Devices: %s", deviceList)

	i := protocol.DeviceState{
		Info: dls.Devices[0],
		Config: &protocol.DeviceConfig{
			SampleRate: 3e6,
		},
	}

	i.Config.RXC = append(i.Config.RXC,
		&protocol.ChannelConfig{
			CenterFrequency: 102.9e6,
			NormalizedGain:  0.85,
			Antenna:         "LNAW",
		})

	deviceProv, _ := json.MarshalIndent(i, "", "   ")
	log.Info("Provisioning Device: %s", deviceProv)

	dinf, err := client.Provision(ctx, &i)
	if err != nil {
		log.Fatal(err)
	}

	deviceInfo, _ := json.MarshalIndent(dinf, "", "   ")
	log.Info("Device Info: %s", deviceInfo)
}
