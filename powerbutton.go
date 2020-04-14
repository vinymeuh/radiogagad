// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/vinymeuh/chardevgpio"
)

type PowerButton struct {
	Chip  string           `yaml:"chip"`
	Lines PowerButtonLines `yaml:"lines"`
}

type PowerButtonLines struct {
	BootOk       int `yaml:"boot_ok"`
	Shutdown     int `yaml:"shutdown"`
	SoftShutdown int `yaml:"soft_shutdown"`
}

func (pb PowerButton) start(msgch chan string, chip chardevgpio.Chip) {
	// set BootOk to High to stop power button flashes
	lineBootOk, err := chip.RequestOutputLine(pb.Lines.BootOk, 1, "bootok")
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup line BootOk: %v", err)
		return
	}
	defer lineBootOk.Close()

	// set SoftShutdown to Low
	lineSoftShutdown, err := chip.RequestOutputLine(pb.Lines.SoftShutdown, 0, "softshutdown")
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup line SoftShutdown: %v", err)
		return
	}
	defer lineSoftShutdown.Close()

	// EventLineWatcher waits for the shutdown button to fire
	watcher, err := chardevgpio.NewEventLineWatcher()
	if err != nil {
		msgch <- fmt.Sprintf("chardevgpio.NewEventLineWatcher: %v", err)
		return
	}
	defer watcher.Close()

	if err := watcher.AddEvent(chip, pb.Lines.Shutdown, "shutdown", chardevgpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup line Shutdown: %v", err)
		return
	}

	if err := watcher.Wait(func(chardevgpio.GPIOEventData) {}); err != nil {
		msgch <- fmt.Sprintf("watcher.Wait: %v", err)
		return
	}

	// Shutown sequence
	msgch <- "Shutdown requested by power button"
	// TODO: display message on screen
	// Set SoftShutdown to high for 1 second to trigger poweroff
	lineSoftShutdown.SetValue(1)
	time.Sleep(1 * time.Second)
	lineSoftShutdown.SetValue(0)
}
