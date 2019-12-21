// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"periph.io/x/periph/conn/gpio"
)

// powerButton is responsible to start a system shutdown when button is fired
func powerButton(msgch chan string, pinBootOk, pinShutdown, pinSoftShutdown gpio.PinIO) {
	if err := pinBootOk.Out(gpio.High); err != nil {
		msgch <- fmt.Sprintf("Failed to setup pinBootOk: %v", err)
		return
	}

	if err := pinShutdown.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup pinShutdown: %v", err)
		return
	}

	if err := pinSoftShutdown.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup pinSoftShutdown: %v", err)
		return
	}

	// poweroff for Alpine Linux, shutdown for Raspbian
	var CmdShutdown string
	if _, err := os.Stat("/sbin/shutdown"); os.IsNotExist(err) {
		CmdShutdown = "poweroff"
	} else {
		CmdShutdown = "/sbin/shutdown -h -P now"
	}
	msgch <- fmt.Sprintf("Shutdown command is '%s'", CmdShutdown)

	go func() {
		for pinShutdown.WaitForEdge(-1) {
			msgch <- "Shutdown requested by power button"
			cmdA := strings.Split(CmdShutdown, " ")
			cmd := exec.Command(cmdA[0], cmdA[1:]...)
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run()
		}
	}()

	go func() {
		for pinSoftShutdown.WaitForEdge(-1) {
			msgch <- "Soft shutdown requested by power button"
			cmdA := strings.Split(CmdShutdown, " ")
			cmd := exec.Command(cmdA[0], cmdA[1:]...)
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run()
		}
	}()

	msgch <- "Power button successfully setup"
}
