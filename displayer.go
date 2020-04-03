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
	Chip    string       `yaml:"chip"`
	Lines   DisplayLines `yaml:"lines"`
	Width   int          `yaml:"line_width"`
	display *weh001602a.Display
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
	txt    string
	length int
	start  int
	end    int
}

func (l *line) reset() {
	l.txt = ""
	l.length = 0
	l.start = 0
	l.end = 0
}

func (l *line) scroll(lineWidth int) {
	l.start++
	l.end++
	if l.end > l.length {
		l.start = 1
		l.end = lineWidth
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
	clrscr.Add(1)

	d.display.CreateChar(cPause, glyphPause)
	d.display.CreateChar(cStop, glyphStop)

	// greeting message
	d.display.Clear()
	d.display.Line1().Write(d.fmtCentred("Hello"))
	time.Sleep(2 * time.Second)

	// main
	ticker := time.NewTicker(400 * time.Millisecond)
	go func() {
		for {
			select {
			//-- refresh screen --//
			case <-ticker.C:
				if d.refresh == true {
					if d.line1.length > 0 {
						d.display.Line1().Write(d.line1.txt[d.line1.start:d.line1.end])
						d.line1.scroll(d.Width)
					}

					if d.line2.length > 0 {
						d.display.Line2().Write(d.line2.txt[d.line2.start:d.line2.end])
						d.line2.scroll(d.Width)
					}
				}
			//-- shutdown --//
			case <-stopscr:
				d.shutdown()
				clrscr.Done()
				return

			//-- display info retrieves from mpdinfo --//
			case data := <-mpdinfo:
				switch data.State {
				case "play":
					d.refresh = true
					// extracts informations to be displayed
					switch data.File[0:4] {
					case "http":
						msgch <- fmt.Sprintf("Playing radio='%s', title='%s'", data.Name, data.Title)
						d.line1.txt = data.Name
						d.line2.txt = data.Title
					default:
						msgch <- fmt.Sprintf("Playing artist='%s', album='%s', title='%s', %d/%d\n",
							data.Artist, data.Album, data.Title, data.Song+1, data.PlaylistLength)
						d.line1.txt = data.Artist
						d.line2.txt = data.Title
					}
					// write line1 & line2
					d.display.Clear()

					if len(d.line1.txt) < d.Width-1 {
						d.line1.length = 0
						d.display.Line1().Write(d.fmtCentred(d.line1.txt))
					} else {
						d.line1.txt = fmt.Sprintf("%s                %s", d.line1.txt, d.line1.txt[0:d.Width-1])
						d.line1.length = len(d.line1.txt)
						d.line1.start = 0
						d.line1.end = d.Width - 1
						// display delayed to next tick
					}

					if len(d.line2.txt) < d.Width-1 {
						d.line2.length = 0
						d.display.Line2().Write(d.fmtCentred(d.line2.txt))
					} else {
						d.line2.txt = fmt.Sprintf("%s                %s", d.line2.txt, d.line2.txt[0:d.Width-1])
						d.line2.length = len(d.line2.txt)
						d.line2.start = 0
						d.line2.end = d.Width - 1
						// display delayed to next tick
					}
				case "pause":
					d.refresh = false
					msgch <- "Player paused"
					d.line1.reset()
					d.line2.reset()
					d.display.Clear()
					d.display.Line1().Write(d.fmtCentred("Pause"))

				default:
					d.refresh = false
					msgch <- "Player stopped"
					d.line1.reset()
					d.line2.reset()
					d.display.Clear()
					d.display.Line1().Write(d.fmtCentred("Stop"))
				}
			}
		}
	}()
}

func (d Displayer) shutdown() {
	d.display.Clear()
	d.display.Line1().Write(d.fmtCentred("Bye Bye"))
	time.Sleep(2 * time.Second)
	d.display.Clear()
	d.display.Off()
}

func (d Displayer) fmtCentred(txt string) string {
	padLen := int((d.Width - len(txt)) / 2)
	return fmt.Sprintf("%*s%s%*s", padLen, "", txt, padLen, "")[0 : d.Width-1]
}
