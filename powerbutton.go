// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/vinymeuh/chardevgpio"
)

// any error while initializing a GPIO pin is fatal and cause a stop of the whole program
func powerButton(chip chardevgpio.Chip, pinBootOk int, pinShutdown int, pinSoftShutdown int, logmsg chan string) {
	// set pinBootOk to High to stop power button flashes
	lineBootOk, err := chip.RequestOutputLine(pinBootOk, 1, "bootok")
	if err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to setup line BootOk: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
		return
	}
	defer lineBootOk.Close()

	// initialize pinSoftShutdown to Low
	lineSoftShutdown, err := chip.RequestOutputLine(pinSoftShutdown, 0, "softshutdown")
	if err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to setup line SoftShutdown: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
		return
	}
	defer lineSoftShutdown.Close()

	// create an EventLineWatcher for pinShutdown
	watcher, err := chardevgpio.NewEventLineWatcher()
	if err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to create EventLineWatcher: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
		return
	}
	defer watcher.Close()

	if err := watcher.AddEvent(chip, pinShutdown, "shutdown", chardevgpio.RisingEdge); err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to setup line Shutdown: %v", err)
		os.Exit(1)
		return
	}

	// block waiting for the button triggering
	if err := watcher.Wait(func(chardevgpio.GPIOEventData) {}); err != nil {
		logmsg <- fmt.Sprintf("Fatal error, failed to wait forbutton triggering: %v", err)
		os.Exit(1)
		return
	}

	// power off sequence
	logmsg <- "Poweroff requested"
	// TODO: display message on screen
	// Set SoftShutdown to high for 1 second to trigger poweroff
	lineSoftShutdown.SetValue(1)
	time.Sleep(1 * time.Second)
	lineSoftShutdown.SetValue(0)
}
