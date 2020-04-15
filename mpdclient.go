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

// mpdInfo is the format of messages send by MPDFetcher to Display
type mpdInfo struct {
	*mpd.Status
	*mpd.CurrentSong
}

// fetcher retrieves messages from the MPD daemon and writes them in a channel as a MPDInfo structure
func (c MPDClient) fetcher(mpdinfo chan mpdInfo, msgch chan string) {
	previous := mpdInfo{Status: &mpd.Status{}, CurrentSong: &mpd.CurrentSong{}}
	for {
		mpc, err := mpd.NewClient(c.Server)
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
		msgch <- "MPD server connection closed"
	}
}

// starter loads and starts playing a playlist
// In case of the MPD server is not responding, retry for a maximum of
// startupMaxRetries spaced by startupRetryDelaySeconds seconds.
func (c MPDClient) starter(msgch chan string) {
	var (
		mpc *mpd.Client
		err error
	)

	retry := 1
	for {
		if mpc, err = mpd.NewClient(c.Server); err == nil {
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

	status, err := mpc.Status()
	if err != nil {
		msgch <- fmt.Sprintf("Unable to retrieve MPD state, playlists load aborted (%s)", err)
		return
	}
	if status.State == "stop" {
		msgch <- fmt.Sprintf("MPD playback is stopped, try to start it")
		for _, playlist := range c.StartupPlaylists {
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
}
