# radiogagad - the daemon inside my RaspDAC

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/vinymeuh/radiogagad.svg)](https://github.com/vinymeuh/radiogagad/releases/latest)
[![Build Status](https://travis-ci.org/vinymeuh/radiogagad.svg?branch=master)](https://travis-ci.org/vinymeuh/radiogagad)
[![codecov](https://codecov.io/gh/vinymeuh/radiogagad/branch/master/graph/badge.svg)](https://codecov.io/gh/vinymeuh/radiogagad)
[![Go Report Card](https://goreportcard.com/badge/github.com/vinymeuh/radiogagad)](https://goreportcard.com/report/github.com/vinymeuh/radiogagad)

## Build and install

First you need a [Go](https://golang.org/dl/) distribution. Then on the build host, targeting a Raspberry Pi 3

```
make buildarm7
```

Install by simply copy the binary under ```/usr/local/bin``` and setup the service for the service manager used by the distribution

* For systemd: [radiogagad.service](https://github.com/vinymeuh/radiogagad/blob/master/radiogagad.service.systemd)
* For OpenRC: [radiogagad.service](https://github.com/vinymeuh/radiogagad/blob/master/radiogagad.service.openrc)

## Configuration

Configuration is loaded from ```/etc/radiogagad.yml``` if file exists. See [radiogagad.yml.template](radiogagad.yml.template) for configuration variables names and their default values.

## Inspirations, links and references

### RaspDAC pinout

![RaspDAC pinout](https://github.com/vinymeuh/radiogagad/blob/master/assets/audiophonics-i-sabre-v4-dac-es9023-tcxo-raspberry-pi-3-b-pi-3-b-pi-2-a-b-i2s.jpg)

![RaspDAC pinout schema](https://github.com/vinymeuh/radiogagad/blob/master/assets/I-SABRE-V3_FR_1_1.jpg)

### Power Button

1. BOOT OK pin must be set to high to stop power button flashes
2. Soft Shutdown pin must be set to low
3. Wait for rising edge on Shutdown pin
4. When button fired, set Soft Shutdown pin to high for 1 second then low to trigger the power off

Examples:

* [Implementation by Audiophonics](https://github.com/audiophonics/Raspberry-pwr-management)
* [Implementation by fengalin](https://github.com/fengalin/raspdac-on-osmc/tree/master/power/sbin)

### MPD

* [MPD Protocol Specification](https://www.musicpd.org/doc/html/protocol.html)

### Winstar OLED WEH001602A

* [HD44780U Hitachi Datasheet](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf)
* [Winstar_GraphicOLED.py](https://github.com/dhrone/Raspdac-Display/blob/master/Winstar_GraphicOLED.py)
* Wikipedia article for [Hitachi HD44780 LCD controller](https://en.wikipedia.org/wiki/Hitachi_HD44780_LCD_controller)
* [Adafruit_CharacterOLED](https://github.com/ladyada/Adafruit_CharacterOLED)
* [Interfacing 16x2 LCD with Raspberry Pi using GPIO & Python](http://www.rpiblog.com/2012/11/interfacing-16x2-lcd-with-raspberry-pi.html)
