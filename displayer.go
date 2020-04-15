// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/vinymeuh/chardevgpio"
	"github.com/vinymeuh/radiogagad/weh001602a"
)

const (
	displayerLineWidth = 16
)

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

// Displayer manages the display, mainly showing MPD messages received from MPDClient.fetcher
type Displayer struct {
	display *weh001602a.Display // useful to commands impacting both lines (Clear, Off)
	line1   line
	line2   line
	refresh bool // controls screen refresh
}

type line struct {
	disp      *weh001602a.Display
	posCursor func() *weh001602a.Display
	width     int
	txt       string
	length    int
	start     int
	end       int
}

func (l *line) setTxt(txt string) {
	if len(txt) < l.width-1 {
		padding := int((l.width - len(txt)) / 2)
		l.txt = fmt.Sprintf("%*s%s%*s", padding, "", txt, padding, "")[0 : l.width-1]
	} else {
		l.txt = fmt.Sprintf("%s                %s", txt, txt[0:l.width-1])
	}
	l.length = len(l.txt)
	l.start = 0
	l.end = l.width - 1
}

func (l *line) refresh() {
	if l.length == l.width {
		l.posCursor().Write(l.txt)
		return
	}
	// scrolling
	l.posCursor().Write(l.txt[l.start:l.end])
	l.start++
	l.end++
	if l.end > l.length {
		l.start = 0
		l.end = l.width - 1
	}
}

func displayer(chip chardevgpio.Chip, pinRS int, pinE int, pinDB4 int, pinDB5 int, pinDB6 int, pinDB7 int, mpdinfo chan mpdInfo, stopscr chan struct{}, clrscr *sync.WaitGroup, logmsg chan string) {
	var err error

	// initialize display
	var d Displayer
	d.display, err = weh001602a.NewDisplay(chip, pinRS, pinE, pinDB4, pinDB5, pinDB6, pinDB7)
	if err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to setup weh001602a display: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}

	d.line1.width = displayerLineWidth
	d.line1.disp = d.display
	d.line1.posCursor = d.display.Line1

	d.line2.width = displayerLineWidth
	d.line2.disp = d.display
	d.line2.posCursor = d.display.Line2

	clrscr.Add(1)

	d.display.CreateChar(cPause, glyphPause)
	d.display.CreateChar(cStop, glyphStop)

	// greeting message
	d.display.Clear()
	d.line1.setTxt("Hello")
	d.line1.refresh()
	time.Sleep(2 * time.Second)

	// main
	ticker := time.NewTicker(400 * time.Millisecond)
	go func() {
		for {
			select {
			//-- refresh screen --//
			case <-ticker.C:
				if d.refresh == true {
					d.line1.refresh()
					d.line2.refresh()
				}

			//-- retrieves info from mpdinfo --//
			case data := <-mpdinfo:
				switch data.State {
				case "play":
					d.refresh = true
					// extracts informations to be displayed
					switch data.File[0:4] {
					case "http":
						logmsg <- fmt.Sprintf("Playing radio='%s', title='%s'", data.Name, data.Title)
						d.line1.setTxt(data.Name)
						d.line2.setTxt(data.Title)
					default:
						logmsg <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						d.line1.setTxt(data.Artist)
						d.line2.setTxt(data.Title)
					}
					// display refresh is delayed to next tick

				case "pause":
					logmsg <- "Player paused"
					d.line1.setTxt("Pause")
					d.line2.setTxt("")

				default:
					logmsg <- "Player stopped"
					d.line1.setTxt("Stop")
					d.line2.setTxt("")
				}

			//-- shutdown requested --//
			case <-stopscr:
				d.refresh = false
				d.display.Clear()
				d.line1.setTxt("Bye Bye")
				d.line1.refresh()
				time.Sleep(2 * time.Second)
				d.display.Clear()
				d.display.Off()
				clrscr.Done()
				return
			}
		}
	}()
}
