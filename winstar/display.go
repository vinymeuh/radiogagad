// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

// This file contains the public API of the driver
//  - the singleton Display (!!no thread safe!!)
//  - high level commands to control display and write text
//
// Pinout is hard coded for the I-Sabre V3 DAC from Audiophonics

package winstar

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	_RS  = "GPIO7"
	_ES  = "GPIO8"
	_DB4 = "GPIO25"
	_DB5 = "GPIO24"
	_DB6 = "GPIO23"
	_DB7 = "GPIO27"
)

// LineNumber is used by WriteAt() to select the line where to write
type LineNumber int

const (
	// Line1 to select first line of the display
	Line1 LineNumber = iota
	// Line2 to select second line of the display
	Line2
)

var display *weh001602a

// Display returns the initialized driver
func Display() *weh001602a {
	if display == nil {
		host.Init()
		display = &weh001602a{
			rs: gpioreg.ByName(_RS),
			es: gpioreg.ByName(_ES),
			data: [4]gpio.PinIO{
				gpioreg.ByName(_DB4),
				gpioreg.ByName(_DB5),
				gpioreg.ByName(_DB6),
				gpioreg.ByName(_DB7),
			},
		}
		display.initialize()
	}
	return display
}

/*********************/
/** Display Control **/
/*********************/

// Clear display writes space code 20H into all DDRAM addresses.
// It then sets DDRAM address 0 into the address counter,
// and returns the display to its original status if it was shifted.
func (w *weh001602a) Clear() {
	w.sendCommand(instrClearDisplay)
	time.Sleep(10 * time.Millisecond)
}

// DisplayOn turns the display on
func (w *weh001602a) DisplayOn() {
	w.instrDisplayControl = w.instrDisplayControl | displayOn
	fmt.Printf("%-16s %08b (%08b)\n", "DisplayOn", w.instrDisplayControl, displayOn)
	w.sendCommand(w.instrDisplayControl)
}

// DisplayOff turns the display off
func (w *weh001602a) DisplayOff() {
	w.instrDisplayControl = w.instrDisplayControl & displayOff
	fmt.Printf("%-16s %08b (%08b)\n", "DisplayOff", w.instrDisplayControl, displayOff)
	w.sendCommand(w.instrDisplayControl)
}

/********************/
/** Cursor Control **/
/********************/

// Return home sets DDRAM address 0 into the address counter,
// and returns the display to its original status if it was shifted.
// The DDRAM contents do not change.
func (w *weh001602a) Home() {
	w.sendCommand(instrReturnHome)
	//time.Sleep(10 * time.Millisecond)
}

/**********************/
/** Write Characters **/
/**********************/

// Send string to LCD at current position of cursor
func (w *weh001602a) Write(msg string) {
	for _, char := range msg {
		if char > 255 {
			char = 32
		}
		w.sendData(uint8(char))
	}
}

// Send string to LCD at start of one of the lines
func (w *weh001602a) WriteAt(line LineNumber, msg string) {
	switch line {
	case Line1:
		w.sendCommand(addrStartLine1)
	case Line2:
		w.sendCommand(addrStartLine2)
	}
	w.Write(msg)
}
