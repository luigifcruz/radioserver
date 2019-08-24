package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"syscall"

	"github.com/luigifreitas/radioserver"
	"github.com/luigifreitas/radioserver/server"
	"github.com/quan-to/slog"
	"github.com/racerxdl/segdsp/dsp"
)

var log = slog.Scope("RadioServer")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Got panic", r)
			debug.PrintStack()
			os.Exit(255)
		}
	}()

	serverName := "helium"

	log.Info("Server Name: %s", serverName)
	log.Info("Protocol Version: %s", radioserver.ServerVersion.AsString())
	log.Info("SIMD Mode: %s", dsp.GetSIMDMode())

	srv := server.MakeRadioServer(serverName)
	err := srv.Listen(":4050")
	if err != nil {
		log.Error("Error listening: %s", err)
	}
	stop := make(chan bool, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		log.Info("Got SIGTERM! Closing it")
		stop <- true
	}()

	<-stop

	srv.Stop()
	log.Info("Done")
}
