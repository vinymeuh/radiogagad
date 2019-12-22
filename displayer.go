// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"

	"github.com/vinymeuh/radiogagad/weh001602a"
)

// displayer manages the display, mainly showing MPD messages received from MPDFetcher
func displayer(mpdinfo chan mpdInfo, stopscr chan struct{}, clrscr *sync.WaitGroup, msgch chan string, pinRS, pinE, pinD4, pinD5, pinD6, pinD7 gpio.PinIO) {

	display, err := weh001602a.NewDisplay(pinRS, pinE, pinD4, pinD5, pinD6, pinD7)
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup weh001602a display: %v", err)
		return
	}
	clrscr.Add(1)

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				break
			case <-stopscr:
				display.Clear()
				display.Line1().WriteCentred("Bye Bye")
				time.Sleep(2 * time.Second)
				display.Clear()
				clrscr.Done()
				return
			case data := <-mpdinfo:
				switch data.State {
				case "play":
					switch data.File[0:4] {
					case "http":
						msgch <- fmt.Sprintf("Playing radio='%s', title='%s'",
							data.Name, data.Title)
						display.Clear()
						display.Line1().WriteCentred(data.Name)
						display.Line2().WriteCentred(data.Title)
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						display.Clear()
						display.Line1().WriteCentred(data.Artist)
						display.Line2().WriteCentred(data.Title)
					}
				case "pause":
					msgch <- "Player paused"
				default:
					display.Clear()
					msgch <- "Player stopped"
				}
			}
		}
	}()
}
