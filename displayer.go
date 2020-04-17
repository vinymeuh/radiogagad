// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vinymeuh/chardevgpio"
	"github.com/vinymeuh/radiogagad/weh001602a"
)

const (
	displayerLineWidth = 16
	// custom characters index for weh001602a
	cPause = 0
	cStop  = 1
)

// displayCmd is the format of messages send by the main goroutine to displayer
type displayCmd struct {
	state string
	line1 string
	line2 string
}

func displayer(chip chardevgpio.Chip, pinRS int, pinE int, pinDB4 int, pinDB5 int, pinDB6 int, pinDB7 int, dispChan chan displayCmd, logger *log.Logger) {
	var err error

	// initialize display
	display, err := weh001602a.NewDisplay(chip, pinRS, pinE, pinDB4, pinDB5, pinDB6, pinDB7)
	if err != nil {
		logger.Fatalf("Fatal error, failed to setup weh001602a display: %v", err)
	}
	display.CreateChar(cPause, weh001602a.PauseGlyph)
	display.CreateChar(cStop, weh001602a.StopGlyph)

	// variables to control display
	var (
		displayRefresh bool
		displayLine1   displayLine
		displayLine2   displayLine
	)
	displayLine1.posCursor = display.Line1
	displayLine2.posCursor = display.Line2

	// greeting message
	display.Clear()
	displayLine1.setTxt("Hello")
	displayLine2.setTxt("(^_^)")
	time.Sleep(2 * time.Second)

	// main
	ticker := time.NewTicker(400 * time.Millisecond)
	go func() {
		for {
			select {
			//-- refresh screen --//
			case <-ticker.C:
				if displayRefresh == true {
					displayLine1.refresh()
					displayLine2.refresh()
				}

			//-- receive command from main goroutine --//
			case cmd := <-dispChan:
				switch cmd.state {
				case "play":
					displayLine1.setTxt(cmd.line1)
					displayLine2.setTxt(cmd.line2)
					if displayLine1.length > displayerLineWidth || displayLine2.length > displayerLineWidth {
						displayRefresh = true
					}
				case "pause":
					displayLine1.setTxt("Pause")
					displayLine2.setTxt("")
					displayRefresh = false
				case "stop":
					displayLine1.setTxt("Stop")
					displayLine2.setTxt("")
					displayRefresh = false
				case "shutdown":
					displayLine1.setTxt("Bye Bye")
					displayLine2.setTxt("(^_^)")
					time.Sleep(2 * time.Second)
					display.Off()
					os.Exit(0)
				}
			}
		}
	}()
}

type displayLine struct {
	posCursor func() *weh001602a.Display
	txt       string
	length    int
	start     int
	end       int
}

func (l *displayLine) setTxt(txt string) {
	if len(txt) < displayerLineWidth-1 {
		padding := int((displayerLineWidth - len(txt)) / 2)
		l.txt = fmt.Sprintf("%*s%s%*s", padding, "", txt, padding, "")[0 : displayerLineWidth-1]
	} else {
		l.txt = fmt.Sprintf("%s                %s", txt, txt[0:displayerLineWidth-1])
	}
	l.length = len(l.txt)
	l.start = 0
	l.end = displayerLineWidth - 1
	l.refresh()
}

func (l *displayLine) refresh() {
	if l.length == displayerLineWidth {
		l.posCursor().Write(l.txt)
		return
	}
	// scrolling
	l.posCursor().Write(l.txt[l.start:l.end])
	l.start++
	l.end++
	if l.end > l.length {
		l.start = 0
		l.end = displayerLineWidth - 1
	}
}
