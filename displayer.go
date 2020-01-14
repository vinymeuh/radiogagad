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

// number of characters per line
const displayerWidth = 16

// custom characters for weh001602a
const (
	cPause = 0
	cStop  = 1
)

var (
	glyphPause = [8]uint8{
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b00000,
	}
	glyphStop = [8]uint8{
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b00000,
	}
)

// displayer manages the display, mainly showing MPD messages received from MPDFetcher
func displayer(mpdinfo chan mpdInfo, stopscr chan struct{}, clrscr *sync.WaitGroup, msgch chan string, pinRS, pinE, pinD4, pinD5, pinD6, pinD7 gpio.PinIO) {
	// initialize display
	display, err := weh001602a.NewDisplay(pinRS, pinE, pinD4, pinD5, pinD6, pinD7)
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup weh001602a display: %v", err)
		return
	}
	clrscr.Add(1)

	display.CreateChar(cPause, glyphPause)
	display.CreateChar(cStop, glyphStop)

	// greeting message
	display.Clear()
	display.Line1().Write(centred("Hello"))
	time.Sleep(2 * time.Second)

	// lines to display
	var (
		line1Txt   string
		line1Len   int
		line1Start int
		line1End   int
		line2Txt   string
		line2Len   int
		line2Start int
		line2End   int
	)

	// main
	ticker := time.NewTicker(400 * time.Millisecond)
	go func() {
		for {
			select {
			//-- lines scrolling --//
			case <-ticker.C:
				if line1Len > 0 {
					display.Line1().Write(line1Txt[line1Start:line1End])
					line1Start++
					line1End++
					if line1End > line1Len {
						line1Start = 1
						line1End = displayerWidth
					}
				}

				if line2Len > 0 {
					display.Line2().Write(line2Txt[line2Start:line2End])
					line2Start++
					line2End++
					if line2End > line2Len {
						line2Start = 1
						line2End = displayerWidth
					}
				}
			//-- shutdown --//
			case <-stopscr:
				display.Clear()
				display.Line1().Write(centred("Bye Bye"))
				time.Sleep(2 * time.Second)
				display.Clear()
				clrscr.Done()
				return
			//-- display info retrieves from mpdinfo --//
			case data := <-mpdinfo:
				switch data.State {
				case "play":
					// extracts informations to be displayed
					switch data.File[0:4] {
					case "http":
						msgch <- fmt.Sprintf("Playing radio='%s', title='%s'",
							data.Name, data.Title)
						line1Txt = data.Name
						line2Txt = data.Title
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						line1Txt = data.Artist
						line2Txt = data.Title
					}
					// write line1 & line2
					display.Clear()

					if len(line1Txt) < displayerWidth-1 {
						line1Len = 0
						display.Line1().Write(centred(line1Txt))
					} else {
						line1Txt = fmt.Sprintf("%s                %s", line1Txt, line1Txt[0:displayerWidth-1])
						line1Len = len(line1Txt)
						line1Start = 0
						line1End = displayerWidth - 1
						// display delayed to next tick
					}

					if len(line2Txt) < displayerWidth-1 {
						line2Len = 0
						display.Line2().Write(centred(line2Txt))
					} else {
						line2Txt = fmt.Sprintf("%s                %s", line2Txt, line2Txt[0:displayerWidth-1])
						line2Len = len(line2Txt)
						line2Start = 0
						line2End = displayerWidth - 1
						// display delayed to next tick
					}
				case "pause":
					msgch <- "Player paused"
					display.Clear()
					display.Line1().Write(centred("Pause"))
				default:
					msgch <- "Player stopped"
					display.Clear()
					display.Line1().Write(centred("Stop"))
				}
			}
		}
	}()
}

func centred(txt string) string {
	padLen := int((displayerWidth - len(txt)) / 2)
	return fmt.Sprintf("%*s%s%*s", padLen, "", txt, padLen, "")[0 : displayerWidth-1]
}
