// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

type MPDInfo struct {
	*mpd.Status
	*mpd.CurrentSong
}

func MPDFetcher(addr string, mpdinfo chan MPDInfo, msgch chan string) {
	var previous mpd.CurrentSong
	for {
		mpc, err := mpd.Dial(addr)
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
				//@FIXME: error handling
			}

			cs, err := mpc.CurrentSong()
			if err != nil {
				msgch <- fmt.Sprintf("Failed to retrieve MPD CurrentSong: %s", err)
				//@FIXME: error handling
			}

			// pass data to the Displayer
			if cs.Name != previous.Name || cs.Title != previous.Title {
				mpdinfo <- MPDInfo{Status: status, CurrentSong: cs}
				previous = *cs
			}

			// waits for notifications from MPD server
			mpc.Idle("player")
		}
	}
}
