// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	GPIO_BOOTOK        = "GPIO22"
	GPIO_SHUTDOWN      = "GPIO17"
	GPIO_SOFT_SHUTDOWN = "GPIO4"
)

// PowerButton is responsible to start a system shutdown when button is fired
func PowerButton(msgch chan string) {
	if _, err := host.Init(); err != nil {
		msgch <- fmt.Sprintf("Failed to run host.Init(): %v", err)
		return
	}

	pinBootOk := gpioreg.ByName(GPIO_BOOTOK)
	if pinBootOk == nil {
		msgch <- "Failed to find pinBootOk"
		return
	}
	if err := pinBootOk.Out(gpio.High); err != nil {
		msgch <- fmt.Sprintf("Failed to setup pinBootOk: %v", err)
		return
	}

	pinShutdown := gpioreg.ByName(GPIO_SHUTDOWN)
	if pinShutdown == nil {
		msgch <- "Failed to find pinShutdown"
		return
	}
	if err := pinShutdown.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		msgch <- fmt.Sprintf("Failed to setup pinShutdown: %v", err)
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
		msgch <- "Power button successfully setup"
		for pinShutdown.WaitForEdge(-1) {
			msgch <- "Power button fired"
			cmdA := strings.Split(CmdShutdown, " ")
			cmd := exec.Command(cmdA[0], cmdA[1:]...)
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run()
		}
	}()
}
