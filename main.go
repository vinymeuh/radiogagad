// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/vinymeuh/chardevgpio"
)

// variables set at build time
var (
	buildVersion string
	buildDate    string
)

func main() {
	// if requested print only build version then exit
	version := flag.Bool("version", false, "Print version and exit.")
	flag.Parse()
	if *version {
		fmt.Println(buildVersion)
		os.Exit(0)
	}

	// initialize logger
	var logmsg *log.Logger
	logmsg = log.New(os.Stdout, "", 0)
	logmsg.Printf("Starting radiogagad %s built %s using %s (%s/%s)\n", buildVersion, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// load configuration
	config := defaultConfiguration()
	err := config.loadFromFile(confFile)
	if err == nil {
		logmsg.Printf("Using configuration file %s\n", confFile)
	} else {
		if !os.IsNotExist(err) {
			logmsg.Printf("Unable to read configuration file: %v\n", err)
			os.Exit(1)
		}
		logmsg.Printf("No configuration file found, we will use the default values\n")
	}
	logmsg.Printf("Using MPD server address %s\n", config.MPD.Server)

	// initialize GPIO chip
	chip, err := chardevgpio.Open(config.PowerButton.Chip)
	if err != nil {
		logmsg.Printf("Failed to call gpio.Open(\"%s\"): %v", config.PowerButton.Chip, err)
		os.Exit(1)
	}

	// this channel will be used by goroutines to return messages to main
	var logch = make(chan string, 32) // buffered channel can hold up to 32 messages before block

	// this channel is used to notify Displayer before shutting down
	var stopscr = make(chan struct{})
	// this wait group is used for waiting that Displayer clear the screen before exit
	var clrscr sync.WaitGroup

	// launches the goroutine responsible to manage the power button
	go powerButton(chip, config.PowerButton.Lines.BootOk, config.PowerButton.Lines.Shutdown, config.PowerButton.Lines.SoftShutdown, logch)

	// launches the goroutine responsible to start playback of a playlist
	go mpdStarter(config.MPD.Server, config.MPD.StartupPlaylists, logch)

	// launches the goroutine responsible to fetch information from MPD
	var mpdinfo = make(chan mpdInfo, 1) // used to return data from MPDFetcher to main goroutine
	go mpdFetcher(config.MPD.Server, mpdinfo, logch)

	// launches the goroutine which manage the display
	go config.Displayer.start(chip, mpdinfo, stopscr, &clrscr, logch)

	// signal handler for SIGTERM & SIGINT
	var shutdown = make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)

	// main loop
	for {
		select {
		case msg := <-logch:
			logmsg.Println(msg)
		case <-shutdown:
			logmsg.Println("Shutdown requested")
			stopscr <- struct{}{}
			clrscr.Wait()
			os.Exit(0)
		}
	}
}
