// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/vinymeuh/chardevgpio"
	"github.com/vinymeuh/radiogagad/weh001602a"
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
	Chip    string              `yaml:"chip"`
	Lines   DisplayLines        `yaml:"lines"`
	Width   int                 `yaml:"line_width"`
	display *weh001602a.Display // useful to commands impacting both lines (Clear, Off)
	line1   line
	line2   line
	refresh bool // controls screen refresh
}

// DisplayLines is the pinout setup of the display
type DisplayLines struct {
	RS  int `yaml:"rs"`
	E   int `yaml:"e"`
	DB4 int `yaml:"db4"`
	DB5 int `yaml:"db5"`
	DB6 int `yaml:"db6"`
	DB7 int `yaml:"db7"`
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

func (d Displayer) start(chip chardevgpio.Chip, mpdinfo chan mpdInfo, stopscr chan struct{}, clrscr *sync.WaitGroup, msgch chan string) {
	var err error

	// initialize display
	d.display, err = weh001602a.NewDisplay(chip, d.Lines.RS, d.Lines.E, d.Lines.DB4, d.Lines.DB5, d.Lines.DB6, d.Lines.DB7)
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup weh001602a display: %v", err)
		return
	}

	d.line1.width = d.Width
	d.line1.disp = d.display
	d.line1.posCursor = d.display.Line1

	d.line2.width = d.Width
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
						msgch <- fmt.Sprintf("Playing radio='%s', title='%s'", data.Name, data.Title)
						d.line1.setTxt(data.Name)
						d.line2.setTxt(data.Title)
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						d.line1.setTxt(data.Artist)
						d.line2.setTxt(data.Title)
					}
					// display refresh is delayed to next tick

				case "pause":
					msgch <- "Player paused"
					d.line1.setTxt("Pause")
					d.line2.setTxt("")

				default:
					msgch <- "Player stopped"
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
