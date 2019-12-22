// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

// mpdInfo is the format of messages send by MPDFetcher to Displayer
type mpdInfo struct {
	*mpd.Status
	*mpd.CurrentSong
}

// mpdFetcher retrieves messages from the MPD daemon and writes them in a channel as a MPDInfo structure
func mpdFetcher(addr string, mpdinfo chan mpdInfo, msgch chan string) {
	var previous mpd.CurrentSong
	for {
		mpc, err := mpd.NewClient(addr)
		if err != nil {
			msgch <- fmt.Sprintf("MPD server not responding: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		// infinite fetch loop
		for {
			status, err := mpc.Status()
			if err != nil {
				msgch <- fmt.Sprintf("Failed to retrieve MPD Status: %s", err)
				goto ResetConnection
			}

			cs, err := mpc.CurrentSong()
			if err != nil {
				msgch <- fmt.Sprintf("Failed to retrieve MPD CurrentSong: %s", err)
				goto ResetConnection
			}

			// pass data to the Displayer
			if cs.Name != previous.Name || cs.Title != previous.Title {
				mpdinfo <- mpdInfo{Status: status, CurrentSong: cs}
				previous = *cs
			}

			// waits for notifications from MPD server
			mpc.Idle("player")
		}
	ResetConnection:
		mpc.Close()
		msgch <- "MPD server connection closed"
	}
}
