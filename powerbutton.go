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

type PowerButton struct {
	Enabled     bool   `yaml:"enabled"`
	ShutdownCmd string `yaml:"shutdown_cmd"`
	Chip        string `yaml:"chip"`
	Lines       struct {
		BootOk       int `yaml:"boot_ok"`
		Shutdown     int `yaml:"shutdown"`
		SoftShutdown int `yaml:"soft_shutdown"`
	} `yaml:"lines"`
}

func (pb PowerButton) start(msgch chan string, chip chardevgpio.Chip) {
	// set BootOk to High to stop power button flashes
	lineBootOk, err := chip.RequestOutputLine(pb.Lines.BootOk, 1, "bootok")
	if err != nil {
		msgch <- fmt.Sprintf("Failed to setup line BootOk: %v", err)
		return
	}
	defer lineBootOk.Close()

	// set shutdownCmd
	var shutdownCmd string
	if pb.ShutdownCmd != "" {
		shutdownCmd = pb.ShutdownCmd
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

	if err := watcher.AddEvent(chip, pb.Lines.Shutdown, "shutdown", chardevgpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup line Shutdown: %v", err)
		return
	}

	if err := watcher.AddEvent(chip, pb.Lines.SoftShutdown, "softshutdown", chardevgpio.RisingEdge); err != nil {
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
