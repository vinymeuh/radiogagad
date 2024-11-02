// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	mpd "github.com/vinymeuh/go-mpdclient"
)

const (
	mpdStarterMaxRetries = 10
	mpdStarterRetryDelay = 2 * time.Second
)

// mpdStarter loads and starts playing a playlist
// In case of the MPD server is not responding, retry for a maximum of
// startupMaxRetries spaced by startupRetryDelaySeconds seconds.
func mpdStarter(addr string, playlists []string, logger *log.Logger) {
	var (
		mpc *mpd.Client
		err error
	)

	retry := 1
	for {
		if mpc, err = mpd.NewClient(addr); err == nil {
			break
		}
		logger.Printf("MPD server not responding: %s", err)
		logger.Printf("Waits %ds before retry", mpdStarterRetryDelay)
		retry++
		if retry > mpdStarterMaxRetries {
			logger.Printf("Unable to contact MPD server after %d retries, we give up", mpdStarterMaxRetries)
			return
		}
		time.Sleep(mpdStarterRetryDelay)
	}

	status, err := mpc.Status()
	if err != nil {
		logger.Printf("Unable to retrieve MPD state, playlists load aborted (%s)", err)
		return
	}

	if status.State == "stop" {
		logger.Printf("MPD playback is stopped, try to start it")
		for _, playlist := range playlists {
			if err := mpc.Load(playlist); err != nil {
				logger.Printf("Failed to load playlist %s (%s)", playlist, err)
				continue
			}
			if err := mpc.Play(-1); err != nil {
				logger.Printf("Failed to start playing playlist %s: %s", playlist, err)
			} else {
				logger.Printf("Successfully started playing playlist '%s'", playlist)
				return
			}
		}
		logger.Printf("Unable to load ANY playlists")
	}
}
