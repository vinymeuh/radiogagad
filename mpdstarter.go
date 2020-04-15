// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/vinymeuh/radiogagad/mpd"
)

const (
	startupMaxRetries        = 10
	startupRetryDelaySeconds = 2
)

// mpdStarter loads and starts playing a playlist
// In case of the MPD server is not responding, retry for a maximum of
// startupMaxRetries spaced by startupRetryDelaySeconds seconds.
func mpdStarter(addr string, playlists []string, logmsg chan string) {
	var (
		mpc *mpd.Client
		err error
	)

	retry := 1
	for {
		if mpc, err = mpd.NewClient(addr); err == nil {
			break
		}
		logmsg <- fmt.Sprintf("MPD server not responding: %s", err)
		logmsg <- fmt.Sprintf("Waits %ds before retry", startupRetryDelaySeconds)
		retry++
		if retry > startupMaxRetries {
			logmsg <- fmt.Sprintf("Unable to contact MPD server after %d retries, we give up", startupMaxRetries)
			return
		}
		time.Sleep(startupRetryDelaySeconds * time.Second)
	}

	status, err := mpc.Status()
	if err != nil {
		logmsg <- fmt.Sprintf("Unable to retrieve MPD state, playlists load aborted (%s)", err)
		return
	}

	if status.State == "stop" {
		logmsg <- fmt.Sprintf("MPD playback is stopped, try to start it")
		for _, playlist := range playlists {
			if err := mpc.Load(playlist); err != nil {
				logmsg <- fmt.Sprintf("Failed to load playlist %s (%s)", playlist, err)
				continue
			}
			if err := mpc.Play(-1); err != nil {
				logmsg <- fmt.Sprintf("Failed to start playing playlist %s: %s", playlist, err)
			} else {
				logmsg <- fmt.Sprintf("Successfully started playing playlist '%s'", playlist)
				return
			}
		}
		logmsg <- "Unable to load ANY playlists"
	}
}
