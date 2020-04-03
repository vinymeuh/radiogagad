// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/vinymeuh/chardevgpio"
	"gopkg.in/yaml.v2"
)

const confFile = "/etc/radiogagad.yml"

var (
	// variables set at build time
	buildVersion string
	buildDate    string
)

// Config is the format of the application's configuration file
type Config struct {
	MPD         MPDClient `yaml:"mpd"`
	PowerButton `yaml:"powerbutton"`
	Displayer   `yaml:"display"`
}

func main() {
	// logger for main goroutine
	var logmsg *log.Logger
	logmsg = log.New(os.Stdout, "", 0)
	logmsg.Printf("Starting radiogagad %s built %s using %s (%s/%s)\n", buildVersion, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// initialize defaults configuration
	config := Config{
		MPD: MPDClient{Server: "localhost:6600"},
		PowerButton: PowerButton{
			Chip: "/dev/gpiochip0",
			Lines: PowerButtonLines{
				BootOk:       22,
				Shutdown:     17,
				SoftShutdown: 4,
			},
		},
		Displayer: Displayer{
			Chip:  "/dev/gpiochip0",
			Width: 16,
			Lines: DisplayLines{
				RS:  7,
				E:   8,
				DB4: 25,
				DB5: 24,
				DB6: 23,
				DB7: 27,
			},
		},
	}

	// load YAML configuration if exists
	err := config.LoadFromFile(confFile)
	if err != nil {
		if !os.IsNotExist(err) {
			logmsg.Printf("Unable to read configuration file: %v\n", err)
			os.Exit(1)
		}
		logmsg.Printf("No configuration file found, we will use the default values\n")
	}
	logmsg.Printf("Using MPD server address %s\n", config.MPD.Server)

	// this channel will be used by goroutines to return messages to main
	var logch = make(chan string, 32) // buffered channel can hold up to 32 messages before block
	// this channel will be used to exchange data from MPDFetcher to Displayer
	var mpdinfo = make(chan mpdInfo, 1)
	// this channel is used to notify Displayer before shutting down
	var stopscr = make(chan struct{})
	// this wait group is used for waiting that Displayer clear the screen before exit
	var clrscr sync.WaitGroup

	// signal handler for SIGTERM & SIGINT
	var stop = make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)
	go func() {
		_ = <-stop
		// notify Displayer and wait it finished
		stopscr <- struct{}{}
		clrscr.Wait()
		os.Exit(0)
	}()

	// initialize GPIO chip
	chip, err := chardevgpio.Open(config.PowerButton.Chip)
	if err != nil {
		logmsg.Printf("Failed to call gpio.Open(\"%s\"): %v", config.PowerButton.Chip, err)
		os.Exit(1)
	}

	// launches the goroutine responsible to manage the power button
	go config.PowerButton.start(logch, chip)

	// launches the goroutine responsible to start playback of a playlist
	go config.MPD.starter(logch)

	// launches the goroutine responsible to fetch information from MPD
	go config.MPD.fetcher(mpdinfo, logch)

	// launches the goroutine which manage the display
	go config.Displayer.start(chip, mpdinfo, stopscr, &clrscr, logch)

	// main loop waits for messages from goroutines
	for {
		msg := <-logch
		logmsg.Println(msg)
	}
}

// LoadFromFile fills the Config from a YAML file
func (c *Config) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	return err
}
