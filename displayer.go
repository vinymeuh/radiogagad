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
				writeTo(display, "Bye Bye", "")
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
						writeTo(display, data.Name, data.Title)
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						writeTo(display, data.Artist, data.Title)
					}
				case "pause":
					msgch <- "Player paused"
					writeTo(display, "Pause", "")
				default:
					msgch <- "Player stopped"
					writeTo(display, "Stop", "")
				}
			}
		}
	}()
}

func writeTo(d *weh001602a.Display, line1 string, line2 string) {
	d.Clear()

	if len(line1) < weh001602a.Width-1 {
		d.Line1().WriteCentred(line1)
	} else {
		d.Line1().Write(line1)
	}

	if len(line2) < weh001602a.Width-1 {
		d.Line2().WriteCentred(line2)
	} else {
		d.Line2().Write(line2)
	}
}
