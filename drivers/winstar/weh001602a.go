// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

// Driver for the Winstar 16x2 Character OLED WEH001602A restricted to the RaspDAC use case.
// It works in 4-bit read-only mode (RW pin connected to ground)
//
// In this restricted use case I can ignore most of the errors, must works or crash !
// So the code should stay readable and easily comparable with Winstar specifications.
//
// This file contains the heart of the beast:
//  - strucutre definition for the driver and constants
//  - low level functions to send command or data to the hardware

package winstar

import (
	"time"

	"periph.io/x/periph/conn/gpio"
)

const (
	// instructions (with all flags to zero)
	instrClearDisplay        = 0x01
	instrReturnHome          = 0x02
	instrEntryModeSet        = 0x04
	instrDisplayOnOffControl = 0x08
	instrCursorShift         = 0x10
	instrFunctionSet         = 0x20
	instrSetCGRAMaddress     = 0x40
	instrSetDDRAMaddress     = 0x80

	// flags for EntryModeSet
	right   = 0x01 << 1
	left    = 0xff & right
	shift   = 0x01
	noshift = 0x00

	// flags for DisplayControl
	displayOn  = 0x01 << 2
	displayOff = 0xff ^ displayOn
	cursorOn   = 0x01 << 1
	cursorOff  = 0xff ^ cursorOn
	blinkOn    = 0x01
	blinkOff   = 0x00

	// flags for FunctionSet
	dataLength8bit       = 0x01 << 4
	dataLength4bit       = 0xff ^ dataLength8bit
	display2lines        = 0x01 << 3
	display1line         = 0xff ^ display2lines
	font5x10             = 0x01 << 2
	font5x8              = 0xff ^ font5x10
	fontWesternEuropean2 = 0x03
	fontEnglishRussian   = 0x02
	fontWesternEuropean1 = 0x01
	fontEnglishJapanese  = 0x00

	// Character Mode Addressing - start address
	addrStartLine1 = 0x80
	addrStartLine2 = 0xc0
)

type weh001602a struct {
	rs                  gpio.PinIO    // Register Select (High=DATA, Low=Instruction Code)
	es                  gpio.PinIO    // Chip Enable Signal
	data                [4]gpio.PinIO // Data bit 4-7
	instrDisplayControl uint8         //
}

func (w *weh001602a) initialize() {

	w.rs.Out(gpio.Low)
	w.es.Out(gpio.Low)
	for i := 0; i < 4; i++ {
		w.data[i].Out(gpio.Low)
	}
	w.instrDisplayControl = instrDisplayOnOffControl | displayOn // display on, no cursor

	// from HD44780U Hitachi Datasheet - page 46
	// Initializing by Instruction - 4-Bit interface
	time.Sleep(20 * time.Millisecond)  // wait for more than 15ms after VCC rises to 4.5 V
	w.write4bits(0x03)                 // function set (interface is 8 bits long)
	time.Sleep(5 * time.Millisecond)   // wait for more than 4.1ms
	w.write4bits(0x03)                 // function set (interface is 8 bits long)
	time.Sleep(100 * time.Microsecond) // wait for more than 100μs
	w.write4bits(0x03)                 // function set (interface is 8 bits long)
	w.write4bits(0x02)                 // function set, set to 4-bit operation (interface is 8 bits long)

	w.sendCommand(instrFunctionSet | display2lines | fontWesternEuropean1) // specify the number of display lines and font size and table
	w.sendCommand(instrDisplayOnOffControl)                                // display off
	w.sendCommand(instrClearDisplay)                                       // display clear
	time.Sleep(10 * time.Millisecond)                                      // clear display is a long instruction (max 6.2ms)
	w.sendCommand(instrEntryModeSet | right)                               // cursor move to the right, noshift

	w.sendCommand(instrReturnHome)       // set DDRAM address 0 in address counter (not done by display clear ?)
	w.sendCommand(w.instrDisplayControl) // display on
}

func (w *weh001602a) sendCommand(bits uint8) {
	w.write8bits(bits, gpio.Low)
}

func (w *weh001602a) sendData(bits uint8) {
	w.write8bits(bits, gpio.High)
}

// write8bits write 8 bits in 4-bit mode:
//  - character and control data are transferred as pairs of 4-bit "nibbles" on the upper data pins, DB7-DB4.
//  - the four most significant bits (7-4) must be written first, followed by the four least significant bits (3-0).
// rs controls the register selected
//   - gpio.Low -> Instruction Register
//   - gpio.Hihg -> Data Register
func (w *weh001602a) write8bits(bits uint8, rs gpio.Level) {
	w.rs.Out(rs)
	w.write4bits(bits >> 4)
	w.write4bits(bits)
}

// write4bits write 4 bits on the data pins DB7-DB4
func (w *weh001602a) write4bits(bits uint8) {
	for i := uint(0); i < 4; i++ {
		switch (bits >> i) & 0x01 {
		case 0x00:
			w.data[i].Out(gpio.Low)
		case 0x01:
			w.data[i].Out(gpio.High)
		}
	}
	w.enable()
}

// Pulse an enable signal on ES pin
func (w *weh001602a) enable() {
	w.es.Out(gpio.High)
	time.Sleep(100 * time.Nanosecond)
	w.es.Out(gpio.Low)
	time.Sleep(100 * time.Nanosecond)
}
