// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

// Test program for the winstar oled driver
// Build with: GOOS=linux GOARCH=arm GOARM=7 go build

package main

import (
	"time"

	"github.com/vinymeuh/radiogagad/winstar"
)

func main() {
	lcd := winstar.Display()

	lcd.Write("All we hear is radio ga ga")
	lcd.WriteAt(winstar.Line2, "Radio goo goo")
	time.Sleep(3 * time.Second)

	lcd.DisplayOff()
	time.Sleep(3 * time.Second)

	lcd.Home()
	lcd.Clear()
	lcd.WriteAt(winstar.Line1, "Don't stop me now")
	lcd.WriteAt(winstar.Line2, "Have a good time, good time")
	lcd.DisplayOn()
	time.Sleep(3 * time.Second)

	lcd.DisplayOff()
}
