// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package weh001602a

import (
	"time"

	"periph.io/x/periph/conn/gpio"
)

// Display is the driver for the Winstar 16x2 Character OLED WEH001602A in 4-bit read-only mode
type Display struct {
	rs  gpio.PinIO // Register Select (High=DATA, Low=Instruction Code)
	e   gpio.PinIO // Enable Signal
	db4 gpio.PinIO // DB4
	db5 gpio.PinIO // DB5
	db6 gpio.PinIO // DB6
	db7 gpio.PinIO // DB7
}

// NewDisplay returns an initialized display
func NewDisplay(rs, e, db4, db5, db6, db7 gpio.PinIO) (*Display, error) {
	if err := rs.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := e.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := db4.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := db5.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := db6.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := db7.Out(gpio.Low); err != nil {
		return nil, err
	}

	d := Display{rs: rs, e: e, db4: db4, db5: db5, db6: db6, db7: db7}
	err := d.initialize()
	return &d, err
}

// Clear clears entire display and sets DDRAM address 0 into the address counter
func (d *Display) Clear() {
	d.sendCommand(0b00000001)
	time.Sleep(10 * time.Millisecond) // long instruction, max 6.2ms
}

// CreateChar creates a custom character for use on the display.
// Up to eight characters (num 0 to 7) of 5x8 dots can be defined.
// Character appearance is defined by an array of eight bytes, one for each line.
func (d *Display) CreateChar(num uint8, pattern [8]uint8) {
	// set the CGRAM address
	switch num {
	case 0:
		d.sendCommand(0x40)
	case 1:
		d.sendCommand(0x48)
	case 2:
		d.sendCommand(0x50)
	case 3:
		d.sendCommand(0x58)
	case 4:
		d.sendCommand(0x60)
	case 5:
		d.sendCommand(0x68)
	case 6:
		d.sendCommand(0x70)
	case 7:
		d.sendCommand(0x78)
	}
	// write character pattern
	for _, w := range pattern {
		d.sendData(w)
	}
}

// Line1 sets cursor at start of first line
func (d *Display) Line1() *Display {
	d.sendCommand(0x80)
	return d
}

// Line2 sets cursor at start of second line
func (d *Display) Line2() *Display {
	d.sendCommand(0xc0)
	return d
}

// Off turns the display off
func (d *Display) Off() {
	d.sendCommand(0b00001000)
}

// Write writes string translated to Western European font table 2
func (d *Display) Write(txt string) *Display {
	for _, char := range txt {
		// code 32 to 126 from font table matchs exactly ASCII
		switch {
		case char >= 32 && char <= 126:
			d.sendData(uint8(char)) // font table matches exactly ASCII
		case char == 224:
			d.sendData(133) // à
		case char == 226:
			d.sendData(131) // â
		case char == 228:
			d.sendData(132) // ä
		case char == 232:
			d.sendData(138) // è
		case char == 233:
			d.sendData(130) // é
		case char == 234:
			d.sendData(136) // ê
		case char == 235:
			d.sendData(137) // ë
		case char == 238:
			d.sendData(140) // î
		case char == 239:
			d.sendData(139) // ï
		case char == 244:
			d.sendData(148) // ô
		case char == 246:
			d.sendData(149) // ö
		case char == 249:
			d.sendData(151) // ù
		case char == 251:
			d.sendData(150) // û
		case char == 252:
			d.sendData(129) // ü
		case char == 255:
			d.sendData(152) // ÿ
		case char == 231:
			d.sendData(135) // ç
		}
	}
	return d
}

// WriteChar writes character corresponding to code in font table
// Notes that code from 0 to 7 are for custom defined characters
func (d *Display) WriteChar(code uint8) *Display {
	d.sendData(code)
	return d
}

// from HD44780U Hitachi Datasheet - page 46
// Initializing by Instruction - 4-Bit interface
func (d *Display) initialize() error {
	time.Sleep(20 * time.Millisecond) // wait for more than 15ms after VCC rises to 4.5 V

	// boot sequence to switch from 8-bit interface to 4-bit interface
	var bootSequence = []struct {
		cmd  uint8
		wait time.Duration
	}{
		{0b00110000, 5 * time.Millisecond},   // init 1st cycle, wait for more than 4.1ms
		{0b00110000, 100 * time.Microsecond}, // init 2nd cycle, wait for more than 100μs
		{0b00110000, 0},                      // init 3rd cycle
		{0b00100000, 0},                      // switch to 4-bit operation
	}
	for _, bs := range bootSequence {
		d.write4bits(bs.cmd)
		if bs.wait > 0 {
			time.Sleep(bs.wait)
		}
	}

	// finish initialization sending commands in 4-bits mode
	var initSequence = []struct {
		cmd  uint8
		wait time.Duration
	}{
		{0b00101011, 0}, // function set, 4-bit (d4=0), 2 lines (d3=1), 5x8dots (d2=0), Western European font table 2 (d1d0=11)
		{0b00001000, 0}, // display off (d2=0)
		{0b00000110, 0}, // entry mode set, increment/move right (d1=1), noshift (d0=0)
		{0b00000010, 0}, // return home
		{0b00001100, 0}, // display on (d2=1), no cursor (d1=0), no blink (d0=0)
	}
	for _, is := range initSequence {
		d.sendCommand(is.cmd)
		if is.wait > 0 {
			time.Sleep(is.wait)
		}
	}

	return nil
}

// sendCommand call write8bits with rs set to LOW
func (d *Display) sendCommand(bits uint8) {
	d.write8bits(bits, gpio.Low)
}

// sendData call write8bits with rs set to HIGH
func (d *Display) sendData(bits uint8) {
	d.write8bits(bits, gpio.High)
}

// write8bits write 8 bits in 4-bit mode
//  - character and control data are transferred as pairs of 4-bit "nibbles" on the upper data pins, DB7-DB4.
//  - the four most significant bits (7-4) must be written first, followed by the four least significant bits (3-0).
// rs controls the register selected
//   - gpio.Low -> Instruction Register
//   - gpio.Hihg -> Data Register
func (d *Display) write8bits(bits uint8, rs gpio.Level) {
	d.rs.Out(rs)
	d.write4bits(bits)
	d.write4bits(bits << 4)
}

// write4bits write the four most significant bits (4-7) to the data pins DB4 to DB7
func (d *Display) write4bits(bits uint8) {
	// DB4
	switch (bits >> 4) & 0x01 {
	case 0x00:
		d.db4.Out(gpio.Low)
	case 0x01:
		d.db4.Out(gpio.High)
	}
	// DB5
	switch (bits >> 5) & 0x01 {
	case 0x00:
		d.db5.Out(gpio.Low)
	case 0x01:
		d.db5.Out(gpio.High)
	}
	// DB6
	switch (bits >> 6) & 0x01 {
	case 0x00:
		d.db6.Out(gpio.Low)
	case 0x01:
		d.db6.Out(gpio.High)
	}
	// DB7
	switch (bits >> 7) & 0x01 {
	case 0x00:
		d.db7.Out(gpio.Low)
	case 0x01:
		d.db7.Out(gpio.High)
	}

	// pulse an enable signal on pin E
	time.Sleep(50 * time.Microsecond)
	d.e.Out(gpio.High)
	time.Sleep(50 * time.Microsecond)
	d.e.Out(gpio.Low)
	time.Sleep(50 * time.Microsecond)
}
