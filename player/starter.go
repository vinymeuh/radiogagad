// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

const (
	STARTUP_MAX_RETRIES = 10
	STARTUP_RETRY_DELAY = 2
)

// Starter loads and starts playing a playlist
// In case of the MPD server is not responding, retry for a maximum of
// STARTUP_MAX_RETRIES spaced by STARTUP_RETRY_DELAY seconds.
func Starter(address string, playlists []string, msgch chan string) {
	var (
		mpc *mpd.Client
		err error
	)

	retry := 1
	for {
		if mpc, err = mpd.Dial(address); err == nil {
			break
		}
		msgch <- fmt.Sprintf("MPD server not responding: %s", err)
		msgch <- fmt.Sprintf("Waits %ds before retry", STARTUP_RETRY_DELAY)
		retry++
		if retry > STARTUP_MAX_RETRIES {
			msgch <- fmt.Sprintf("Unable to contact MPD server after %d retries, we give up", STARTUP_MAX_RETRIES)
			return
		}
		time.Sleep(STARTUP_RETRY_DELAY * time.Second)
	}

	for _, playlist := range playlists {
		if err := mpc.Load(playlist); err != nil {
			msgch <- fmt.Sprintf("Failed to load playlist %s (%s)", playlist, err)
			continue
		}
		if err := mpc.Play(-1); err != nil {
			msgch <- fmt.Sprintf("Failed to start playing playlist %s: %s", playlist, err)
		} else {
			msgch <- fmt.Sprintf("Successfully started playing playlist '%s'", playlist)
			return
		}
	}
	msgch <- "Unable to load ANY playlists"
}
