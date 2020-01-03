// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.
//
// GOOS=linux GOARCH=arm GOARM=6 go build
package main

import (
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/vinymeuh/radiogagad/weh001602a"
)

func main() {
	host.Init()
	display, _ := weh001602a.NewDisplay(
		gpioreg.ByName("GPIO7"),
		gpioreg.ByName("GPIO8"),
		gpioreg.ByName("GPIO25"),
		gpioreg.ByName("GPIO24"),
		gpioreg.ByName("GPIO23"),
		gpioreg.ByName("GPIO27"),
	)

	// Send some centred test
	display.Clear()
	display.Line1().WriteCentred("Rasbperry Pi")
	display.Line2().WriteCentred(":)")
	time.Sleep(3 * time.Second)

	// Send some left & right justified text
	display.Clear()
	display.Line1().Write("<- left")
	display.Line2().WriteRightAligned("right ->")
	time.Sleep(3 * time.Second)

	// Create & use custom characters
	bell := [8]uint8{
		0b00100,
		0b01110,
		0b01110,
		0b01110,
		0b01110,
		0b11111,
		0b00100,
		0b00000,
	}
	battery := [8]uint8{
		0b01110,
		0b11011,
		0b10001,
		0b10001,
		0b11111,
		0b11111,
		0b11111,
		0b00000,
	}

	display.CreateChar(0, bell)
	display.CreateChar(1, battery)

	// Write custom char
	display.Clear()
	display.Line1().WriteChar(0).Write(" ALARM")
	display.Line2().WriteChar(1).Write(" BATTERY")
	time.Sleep(3 * time.Second)

	// Write characters font table
	for i := 0; i <= 255; i++ {
		display.Clear()
		display.Line1().Write(strconv.Itoa(i))
		display.Line2().WriteChar(uint8(i))
		time.Sleep(1 * time.Second)
	}

	// Shutdown
	display.Off()
}
