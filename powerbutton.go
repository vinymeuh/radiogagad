// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/vinymeuh/chardevgpio"
)

// any error while initializing a GPIO pin is fatal and cause a stop of the whole program
func powerButton(chip chardevgpio.Chip, pinBootOk int, pinShutdown int, pinSoftShutdown int, logger *log.Logger) {
	// set pinBootOk to High to stop power button flashes
	lineBootOk, err := chip.RequestOutputLine(pinBootOk, 1, "bootok")
	if err != nil {
		logger.Fatalf("Fatal error, failed to setup line BootOk: %v", err)
	}
	defer lineBootOk.Close()

	// initialize pinSoftShutdown to Low
	lineSoftShutdown, err := chip.RequestOutputLine(pinSoftShutdown, 0, "softshutdown")
	if err != nil {
		logger.Fatalf("Fatal error, failed to setup line SoftShutdown: %v", err)
	}
	defer lineSoftShutdown.Close()

	// create an EventLineWatcher for pinShutdown
	watcher, err := chardevgpio.NewEventLineWatcher()
	if err != nil {
		logger.Fatalf("Fatal error, failed to create EventLineWatcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.AddEvent(chip, pinShutdown, "shutdown", chardevgpio.RisingEdge); err != nil {
		logger.Fatalf("Fatal error, failed to setup line Shutdown: %v", err)
	}

	// block waiting for the button triggering
	if err := watcher.Wait(func(chardevgpio.GPIOEventData) {}); err != nil {
		logger.Fatalf("Fatal error, failed to wait forbutton triggering: %v", err)
	}

	// power off sequence
	logger.Print("Poweroff requested")
	// TODO: display message on screen
	// Set SoftShutdown to high for 1 second to trigger poweroff
	lineSoftShutdown.SetValue(1)
	time.Sleep(1 * time.Second)
	lineSoftShutdown.SetValue(0)
}
