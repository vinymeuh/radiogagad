// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"sync"
	"time"

	"github.com/vinymeuh/radiogagad/winstar"
)

// Displayer manages the OLED display, mainly showing MPD messages received from MPDFetcher
func Displayer(mpdinfo chan MPDInfo, stopscr chan struct{}, clrscr *sync.WaitGroup, msgch chan string) {
	lcd := winstar.Display()
	clrscr.Add(1)

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				break
			case <-stopscr:
				lcd.Clear()
				lcd.WriteAt(winstar.Line1, "Stopping")
				time.Sleep(2 * time.Second)
				lcd.Clear()
				clrscr.Done()
				return
			case data := <-mpdinfo:
				switch data.State {
				case "play":
					switch data.File[0:4] {
					case "http":
						msgch <- fmt.Sprintf("Playing radio='%s', title='%s'",
							data.Name, data.Title)
						lcd.Clear()
						lcd.WriteAt(winstar.Line1, data.Name)
						lcd.WriteAt(winstar.Line2, data.Title)
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						lcd.Clear()
						lcd.WriteAt(winstar.Line1, data.Artist)
						lcd.WriteAt(winstar.Line2, data.Title)
					}
				case "pause":
					msgch <- "Player paused"
				default:
					lcd.Clear()
					msgch <- "Player stopped"
				}
			}
		}
	}()
}
