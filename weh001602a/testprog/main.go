// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.
//
// GOOS=linux GOARCH=arm GOARM=6 go build
package main

import (
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

	// kaomoji
	display.Clear()
	k1 := `(°_°)ノ`
	display.Line1().WriteCentred(k1)
	time.Sleep(3 * time.Second)

	display.Clear()
	k2 := `(-。-) Zzzz`
	display.Line1().WriteCentred(k2)
	time.Sleep(3 * time.Second)

	// Write character font table
	// for i := 17; i <= 255; i++ {
	// 	display.Clear()
	// 	c := uint(i)
	// 	display.Line1().Write(strconv.Itoa(int(c)))
	// 	display.Line2().Write(string(c))
	// 	time.Sleep(2 * time.Second)
	// }

	display.Off()
}
