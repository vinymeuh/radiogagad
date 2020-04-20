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
	var logger *log.Logger
	logger = log.New(os.Stdout, "", 0)
	logger.Printf("Starting radiogagad %s built %s using %s (%s/%s)", buildVersion, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// load configuration
	config := defaultConfiguration()
	err := config.loadFromFile(confFile)
	if err == nil {
		logger.Printf("Using configuration file %s", confFile)
	} else {
		if !os.IsNotExist(err) {
			logger.Printf("Unable to read configuration file: %v", err)
			os.Exit(1)
		}
		logger.Printf("No configuration file found, we will use the default values")
	}
	logger.Printf("Using MPD server address %s", config.MPD.Server)

	// initialize GPIO chip
	chip, err := chardevgpio.Open(config.Chip.Device)
	if err != nil {
		logger.Printf("Failed to call gpio.Open(\"%s\"): %v", config.Chip.Device, err)
		os.Exit(1)
	}

	// launches the goroutine responsible to manage the power button
	go powerButton(chip, config.Chip.BootOk, config.Chip.Shutdown, config.Chip.SoftShutdown, logger)

	// launches the goroutine responsible to start playback of a playlist
	go mpdStarter(config.MPD.Server, config.MPD.StartupPlaylists, logger)

	// launches the goroutine responsible to fetch information from MPD
	var mpdChan = make(chan mpdInfo, 1) // used to retrieve data from MPDFetcher
	go mpdFetcher(config.MPD.Server, mpdChan, logger)

	// launches the goroutine which manage the display
	var dispChan = make(chan displayCmd, 1) // used to send command to displayer
	go displayer(chip, config.Chip.RS, config.Chip.E, config.Chip.DB4, config.Chip.DB5, config.Chip.DB6, config.Chip.DB7, dispChan, logger)

	// signal handler for SIGTERM & SIGINT
	var shutdown = make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)

	// main loop
	dispcmd := displayCmd{state: "stop"}
	for {
		dispChan <- dispcmd
		select {
		//-- receive informations from mpdFetcher --//
		case mpdinfo := <-mpdChan:
			switch mpdinfo.State {
			case "play":
				dispcmd.state = "play"
				switch mpdinfo.File[0:4] {
				case "http":
					logger.Printf("Play radio='%s', title='%s'", mpdinfo.Name, mpdinfo.Title)
					dispcmd.line1 = mpdinfo.Name
					dispcmd.line2 = mpdinfo.Title
				default:
					logger.Printf("Play artist='%s', album='%s', title='%s', %d/%d", mpdinfo.Artist, mpdinfo.Album, mpdinfo.Title, mpdinfo.Song+1, mpdinfo.PlaylistLength)
					dispcmd.line1 = mpdinfo.Artist
					dispcmd.line2 = mpdinfo.Title
				}
			case "pause":
				logger.Printf("Player paused")
				dispcmd.state = "pause"
			default:
				logger.Printf("Player stopped")
				dispcmd.state = "stop"
			}
		//-- shutdown requested by SIGTERM/SIGINT (displayer will finish the program) --//
		case <-shutdown:
			logger.Println("Shutdown requested")
			dispcmd.state = "shutdown"
		}
	}
}
