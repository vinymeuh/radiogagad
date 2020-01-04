// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.
//
// GOOS=linux GOARCH=arm GOARM=6 go build
package main

import (
	"fmt"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/vinymeuh/radiogagad/weh001602a"
)

func centred(txt string) string {
	padLen := int((weh001602a.Width - len(txt)) / 2)
	return fmt.Sprintf("%*s%s%*s", padLen, "", txt, padLen, "")[0 : weh001602a.Width-1]
}

func rightAligned(txt string) string {
	return fmt.Sprintf("%*s", weh001602a.Width, txt)
}

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
	display.Line1().Write(centred("Rasbperry Pi"))
	display.Line2().Write(centred(":)"))
	time.Sleep(3 * time.Second)

	// Send some left & right justified text
	display.Clear()
	display.Line1().Write("<- left")
	display.Line2().Write(rightAligned("right ->"))
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
	pause := [8]uint8{
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b11011,
		0b00000,
	}
	stop := [8]uint8{
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b11111,
		0b00000,
	}

	display.CreateChar(0, bell)
	display.CreateChar(1, battery)
	display.CreateChar(2, pause)
	display.CreateChar(3, stop)

	// Write custom char
	display.Clear()
	display.Line1().WriteChar(0).Write(" ALARM")
	display.Line2().WriteChar(1).Write(" BATTERY")
	time.Sleep(3 * time.Second)

	display.Clear()
	display.Line1().WriteChar(2).Write(" Pause")
	display.Line2().WriteChar(3).Write(" Stop")
	time.Sleep(3 * time.Second)

	// Write accented characters
	display.Clear()
	display.Line1().Write("àâäéèêëïîôöùûüÿç")
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
