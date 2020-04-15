// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

// mpdInfo is the format of messages send by mpdFetcher to the main goroutine
type mpdInfo struct {
	*mpd.Status
	*mpd.CurrentSong
}

// mpdFetcher retrieves messages from the MPD daemon and writes them in a channel as a MPDInfo structure
func mpdFetcher(addr string, mpdinfo chan mpdInfo, logmsg chan string) {
	previous := mpdInfo{Status: &mpd.Status{}, CurrentSong: &mpd.CurrentSong{}}
	for {
		mpc, err := mpd.NewClient(addr)
		if err != nil {
			logmsg <- fmt.Sprintf("MPD server not responding: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		// infinite fetch loop
		for {
			status, err := mpc.Status()
			if err != nil {
				logmsg <- fmt.Sprintf("Failed to retrieve MPD Status: %s", err)
				goto ResetConnection
			}

			cs, err := mpc.CurrentSong()
			if err != nil {
				logmsg <- fmt.Sprintf("Failed to retrieve MPD CurrentSong: %s", err)
				goto ResetConnection
			}

			current := mpdInfo{Status: status, CurrentSong: cs}

			// pass data to the Displayer
			if current.Status.State != previous.Status.State ||
				current.CurrentSong.Name != previous.CurrentSong.Name ||
				current.CurrentSong.Title != previous.CurrentSong.Title {
				mpdinfo <- current
				previous = current
			}

			// waits for notifications from MPD server
			mpc.Idle("player")
		}
	ResetConnection:
		mpc.Close()
		logmsg <- "MPD server connection closed"
	}
}
