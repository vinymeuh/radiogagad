// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package weh001602a

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
)

const (
	// Width is the number of characters per line
	Width = 16
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

// Write writes text left aligned
func (d *Display) Write(txt string) {
	for _, char := range []rune(txt) {
		d.sendData(ft00(char))
	}
}

// WriteRightAligned writes text right aligned
func (d *Display) WriteRightAligned(txt string) {
	newtxt := fmt.Sprintf("%*s", Width, txt)
	d.Write(newtxt)
}

// WriteCentred writes text centred
func (d *Display) WriteCentred(txt string) {
	padLen := int((Width - len(txt)) / 2)
	newtxt := fmt.Sprintf("%*s%s%*s", padLen, "", txt, padLen, "")[0 : Width-1]
	d.Write(newtxt)
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
		{0b00101000, 0}, // function set, 4-bit (d4=0), 2 lines (d3=1), 5x8dots (d2=0), English Japanese font table (d1d0=00)
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

// mapping function for English Japanese character font table
func ft00(r rune) uint8 {
	// ASCII
	if r >= 32 && r <= 125 {
		return uint8(r)
	}

	switch r {
	case 176:
		return 223 // `°`
	case 12290: // `。`
		return 161
	case 12494: // `ノ`
		return 201
	}

	if r <= 250 {
		return uint8(r)
	}

	return 32
}
