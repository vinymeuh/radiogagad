// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vinymeuh/chardevgpio"
)

// powerButton is responsible to start a system shutdown when button is fired
func powerButton(msgch chan string, chip chardevgpio.Chip) {

	// set BootOk to High to stop power button flashes
	lineBootOk, err := chip.RequestOutputLine(config.PowerButton.Lines.BootOk, 1, "bootok")
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup line BootOk: %v", err)
		return
	}

	// set shutdownCmd
	var shutdownCmd string
	if config.PowerButton.ShutdownCmd != "" {
		shutdownCmd = config.PowerButton.ShutdownCmd
	} else {
		// try to auto detect
		if _, err := os.Stat("/sbin/shutdown"); os.IsNotExist(err) {
			shutdownCmd = "/sbin/poweroff" // Alpine Linux
		} else {
			shutdownCmd = "/sbin/shutdown -h -P now" // Raspbian
		}
	}
	msgch <- fmt.Sprintf("Shutdown command is '%s'", shutdownCmd)

	// EventLineWatcher waits for the shutdown button to fire
	watcher, err := chardevgpio.NewEventLineWatcher()
	if err != nil {
		msgch <- fmt.Sprintf("chardevgpio.NewEventLineWatcher: %v", err)
		return
	}
	defer watcher.Close()

	if err := watcher.AddEvent(chip, config.PowerButton.Lines.Shutdown, "shutdown", chardevgpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup line Shutdown: %v", err)
		return
	}

	if err := watcher.AddEvent(chip, config.PowerButton.Lines.SoftShutdown, "softshutdown", chardevgpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup line SoftShutdown: %v", err)
		return
	}

	if err := watcher.Wait(func(chardevgpio.GPIOEventData) {}); err != nil {
		msgch <- fmt.Sprintf("watcher.Wait: %v", err)
		return
	}
	msgch <- "Shutdown requested by power button"
	cmdA := strings.Split(shutdownCmd, " ")
	cmd := exec.Command(cmdA[0], cmdA[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Run()
}
