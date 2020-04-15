// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

const confFile = "/etc/radiogagad.yml"

type configuration struct {
	MPD         MPDClient `yaml:"mpd"`
	PowerButton `yaml:"powerbutton"`
	Displayer   `yaml:"display"`
}

type MPDClient struct {
	Server           string   `yaml:"host"`
	StartupPlaylists []string `yaml:"startup_playlists"`
}

type PowerButton struct {
	Chip  string           `yaml:"chip"`
	Lines PowerButtonLines `yaml:"lines"`
}

type PowerButtonLines struct {
	BootOk       int `yaml:"boot_ok"`
	Shutdown     int `yaml:"shutdown"`
	SoftShutdown int `yaml:"soft_shutdown"`
}

type DisplayLines struct {
	RS  int `yaml:"rs"`
	E   int `yaml:"e"`
	DB4 int `yaml:"db4"`
	DB5 int `yaml:"db5"`
	DB6 int `yaml:"db6"`
	DB7 int `yaml:"db7"`
}

func defaultConfiguration() configuration {
	return configuration{
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
			Chip: "/dev/gpiochip0",
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
}

func (c *configuration) loadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	return err
}
