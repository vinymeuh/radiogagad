// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/vinymeuh/radiogagad/player"
)

// variables for build time versioning
var (
	Version string
	Build   string
)

func main() {
	logmsg := log.New(os.Stdout, "", 0)
	logmsg.Printf("Starting radiogagad %s built %s using %s (%s/%s)\n", Version, Build, runtime.Version(), runtime.GOOS, runtime.GOARCH)

	server := os.Getenv("RGGD_MPD_SERVER")
	if server == "" {
		server = "localhost:6600"
	}
	logmsg.Printf("Using MPD server address %s\n", server)

	// this channel will be used by goroutines to return messages to main
	var logch = make(chan string, 32) // buffered channel can hold up to 32 messages before block
	// this channel will be used to exchange data from MPDFetcher to Displayer
	var mpdinfo = make(chan player.MPDInfo, 1)
	// this channel is used by notify Displayer before shutting down
	var stopscr = make(chan struct{})
	// this wait group is used for waiting that Displayer clear the screen before exit
	var clrscr sync.WaitGroup

	// signal handler for SIGTERM & SIGINT
	var stop = make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)
	go func() {
		_ = <-stop
		// notify Displayer and wait it finished
		stopscr <- struct{}{}
		clrscr.Wait()
		os.Exit(0)
	}()

	// launches the goroutine responsible for the power button
	go player.PowerButton(logch)

	// launches the goroutine responsible for starting playback of a playlist
	playlists := strings.Split(os.Getenv("RGGD_STARTUP_PLAYLISTS"), ",")
	if len(playlists) > 0 && playlists[0] != "" {
		go player.Starter(server, playlists, logch)
	}

	// launches the goroutines which manage the display
	go player.MPDFetcher(server, mpdinfo, logch)
	go player.Displayer(mpdinfo, stopscr, &clrscr, logch)

	// main loop waits for messages from goroutines
	for {
		msg := <-logch
		logmsg.Println(msg)
	}

}
