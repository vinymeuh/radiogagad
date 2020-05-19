// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	gpio "github.com/vinymeuh/chardevgpio"
)

// any error while initializing a GPIO pin is fatal and cause a stop of the whole program
func powerButton(chip gpio.Chip, pinBootOk int, pinShutdown int, pinSoftShutdown int, logger *log.Logger) {
	// set pinBootOk to High to stop power button flashes
	lineBootOk := gpio.NewHandleRequest([]int{pinBootOk}, gpio.HandleRequestOutput).WithConsumer("bootok").WithDefaults([]int{1})
	if err := chip.RequestLines(lineBootOk); err != nil {
		logger.Fatalf("Fatal error, failed to setup line BootOk: %v", err)
	}
	defer lineBootOk.Close()

	// initialize pinSoftShutdown to Low
	lineSoftShutdown := gpio.NewHandleRequest([]int{pinSoftShutdown}, gpio.HandleRequestOutput).WithConsumer("softshutdown").WithDefaults([]int{0})
	if err := chip.RequestLines(lineSoftShutdown); err != nil {
		logger.Fatalf("Fatal error, failed to setup line SoftShutdown: %v", err)
	}
	defer lineSoftShutdown.Close()

	// create a LineWatcher for pinShutdown
	watcher, err := gpio.NewLineWatcher()
	if err != nil {
		logger.Fatalf("Fatal error, failed to create LineWatcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(chip, pinShutdown, gpio.RisingEdge, "shutdown"); err != nil {
		logger.Fatalf("Fatal error, failed to setup line Shutdown: %v", err)
	}

	// block waiting for the button triggering
	if _, err := watcher.Wait(); err != nil {
		logger.Fatalf("Fatal error, failed to wait for button triggering: %v", err)
	}

	// power off sequence
	logger.Print("Poweroff requested")
	// TODO: display message on screen
	// Set SoftShutdown to high for 1 second to trigger poweroff
	lineSoftShutdown.Write(1)
	time.Sleep(1 * time.Second)
	lineSoftShutdown.Write(0)
}
