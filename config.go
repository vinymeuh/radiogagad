// Copyright 2020 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config is the format of the application's configuration file
type Config struct {
	MPD struct {
		Server string `yaml:"host"`
	} `yaml:"mpd"`
	PowerButton struct {
		Enabled     bool   `yaml:"enabled"`
		ShutdownCmd string `yaml:"shutdown_cmd"`
		Chip        string `yaml:"chip"`
		Lines       struct {
			BootOk       int `yaml:"boot_ok"`
			Shutdown     int `yaml:"shutdown"`
			SoftShutdown int `yaml:"soft_shutdown"`
		} `yaml:"lines"`
	} `yaml:"powerbutton"`
	Display struct {
		Lines struct {
			RS  int `yaml:"rs"`
			E   int `yaml:"e"`
			DB4 int `yaml:"db4"`
			DB5 int `yaml:"db5"`
			DB6 int `yaml:"db6"`
			DB7 int `yaml:"db7"`
		} `yaml:"lines"`
	} `yaml:"display"`
}

// NewConfig creates a new Config with default values
func NewConfig() Config {
	var c Config
	c.MPD.Server = "localhost:6600"

	c.PowerButton.Enabled = true
	c.PowerButton.Chip = "/dev/gpiochip0"
	c.PowerButton.Lines.BootOk = 22
	c.PowerButton.Lines.Shutdown = 17
	c.PowerButton.Lines.SoftShutdown = 4

	c.Display.Lines.RS = 7
	c.Display.Lines.E = 8
	c.Display.Lines.DB4 = 25
	c.Display.Lines.DB5 = 24
	c.Display.Lines.DB6 = 23
	c.Display.Lines.DB7 = 27
	return c
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
