// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.
//
// GOOS=linux GOARCH=arm GOARM=6 go build
package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/vinymeuh/radiogagad/weh001602a"
)

var display *weh001602a.Display

const displayerWidth = 16

func centred(txt string) string {
	padLen := int((displayerWidth - len(txt)) / 2)
	return fmt.Sprintf("%*s%s%*s", padLen, "", txt, padLen, "")[0 : displayerWidth-1]
}

func rightAligned(txt string) string {
	return fmt.Sprintf("%*s", displayerWidth, txt)
}

func writeAccentedCharacters() {
	display.Clear()
	display.Line1().Write("àâäéèêëïîôöùûüÿç")
	display.Line2().Write("ñØø")
	time.Sleep(3 * time.Second)
}

func writeCustomCharacters() {
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
	time.Sleep(2 * time.Second)

	display.Clear()
	display.Line1().WriteChar(2).Write(" Pause")
	display.Line2().WriteChar(3).Write(" Stop")
	time.Sleep(2 * time.Second)
}

func writeFontTable() {
	for i := 0; i <= 255; i++ {
		display.Clear()
		display.Line1().Write(strconv.Itoa(i))
		display.Line2().WriteChar(uint8(i))
		time.Sleep(1 * time.Second)
	}
}

func writeWithScrolling() {
	display.Clear()
	var w sync.WaitGroup
	w.Add(1)
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		display.Line1().Write("Scrolling")
		msg2 := "I'm scrolling from right to left"
		line2 := fmt.Sprintf("%s                %s", msg2, msg2)
		l2 := len(line2)
		s2 := 0
		e2 := displayerWidth - 1
		tick := 0
		for {
			select {
			case <-ticker.C:
				if tick == 60 {
					w.Done()
					return
				}
				tick++
				display.Line2().Write(line2[s2:e2])
				s2 = (s2 + 1) % l2
				e2 = s2 + (displayerWidth - 1)
				if e2 > l2 {
					e2 = l2
				}
			}
		}
	}()
	w.Wait()
}

type line struct {
	txt    string
	isLong bool
}

func main() {
	host.Init()
	display, _ = weh001602a.NewDisplay(
		gpioreg.ByName("GPIO7"),
		gpioreg.ByName("GPIO8"),
		gpioreg.ByName("GPIO25"),
		gpioreg.ByName("GPIO24"),
		gpioreg.ByName("GPIO23"),
		gpioreg.ByName("GPIO27"),
	)

	// Send some centred test
	display.Clear()
	display.Line1().Write(centred("Raspberry Pi"))
	display.Line2().Write(centred(":)"))
	time.Sleep(2 * time.Second)

	// Send some left & right justified text
	display.Clear()
	display.Line1().Write("<- left")
	display.Line2().Write(rightAligned("right ->"))
	time.Sleep(2 * time.Second)

	writeAccentedCharacters()
	writeCustomCharacters()
	//writeFontTable()
	writeWithScrolling()

	// Shutdown
	display.Clear()
	display.Line1().Write("Shutdown")
	time.Sleep(2 * time.Second)
	display.Clear()
	display.Off()
}
