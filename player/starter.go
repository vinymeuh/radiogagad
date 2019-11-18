// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

const (
	startupMaxRetries        = 10
	startupRetryDelaySeconds = 2
)

// Starter loads and starts playing a playlist
// In case of the MPD server is not responding, retry for a maximum of
// startupMaxRetries spaced by startupRetryDelaySeconds seconds.
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
		msgch <- fmt.Sprintf("Waits %ds before retry", startupRetryDelaySeconds)
		retry++
		if retry > startupMaxRetries {
			msgch <- fmt.Sprintf("Unable to contact MPD server after %d retries, we give up", startupMaxRetries)
			return
		}
		time.Sleep(startupRetryDelaySeconds * time.Second)
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
