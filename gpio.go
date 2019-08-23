// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package gpio provides utilities to query and dink with general purpose i/o
// pins.
package gpio

import (
	"fmt"
	"os"
	"strconv"
)

type Pin struct {
	Gpio    int
	Name    string
	Default string
}

type GpioAliasMap map[string]string
type PinMap map[string]*Pin

type Chip struct {
	// Chip has GPIOs base through base + count.
	Base, Count Pin
	// Value of compatible=XXX node in DTS file for this GPIO chip.
	Compatible map[string]bool
}

var Aliases GpioAliasMap
var Pins PinMap

// File prefix for testing w/o proper sysfs.
var prefix string

func SetDebugPrefix(p string) { prefix = p }

var GpioBankToBase = map[string]int{
	"gpio0": 0,
	"gpio1": 32,
	"gpio2": 64,
	"gpio3": 96,
	"gpio4": 128,
	"gpio5": 160,
	"gpio6": 192,
}

var GpioPinMode = map[string]string{
	"output-high": "high",
	"output-low":  "low",
	"input":       "in",
}

func (p *Pin) Export() (err error) {
	fn := prefix + "/sys/class/gpio/export"
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%d\n", p.Gpio)
	return
}

func (p *Pin) IsExported() (x bool) {
	fn := fmt.Sprintf(prefix+"/sys/class/gpio/gpio%d/value", p.Gpio)
	_, err := os.Stat(fn)
	if err != nil {
		return false
	}
	return true
}

func (p *Pin) Open(name string) (f *os.File, fn string, err error) {
	fn = fmt.Sprintf(prefix+"/sys/class/gpio/gpio%d/%s", p.Gpio, name)
	f, err = os.OpenFile(fn, os.O_RDWR, 0)
	return
}

func (p *Pin) Direction() (dir string, err error) {
	f, _, err := p.Open("direction")
	if err != nil {
		return
	}
	defer f.Close()

	_, err = fmt.Fscanf(f, "%s\n", &dir)

	return
}

// "direction" ... reads as either "in" or "out". This value may
// 	normally be written. Writing as "out" defaults to
// 	initializing the value as low. To ensure glitch free
// 	operation, values "low" and "high" may be written to
// 	configure the GPIO as an output with that initial value.
func (p *Pin) SetDirection(dir string) (err error) {
	f, _, err := p.Open("direction")
	if err != nil {
		return
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", dir)
	return
}

func (p *Pin) SetValue(v bool) (err error) {
	f, _, err := p.Open("value")
	if err != nil {
		return
	}
	defer f.Close()
	x := 0
	if v {
		x = 1
	}
	_, err = fmt.Fprintf(f, "%d\n", x)
	return
}

func (p *Pin) Value() (v bool, err error) {
	f, _, err := p.Open("value")
	if err != nil {
		return
	}
	defer f.Close()
	x := 0
	_, err = fmt.Fscanf(f, "%d\n", &x)
	if x != 0 {
		v = true
	}
	return
}

func (p *Pin) String() string {
	return fmt.Sprintf("Gpio: %d (%s)", p.Gpio, p.Name)
}

func (p *Pin) SetDefault() (err error) {
	return p.SetDirection(p.Default)
}

func NewPin(name, mode, bank, index string) (err error) {
	i, _ := strconv.Atoi(index)
	p := &Pin{Gpio: GpioBankToBase[bank] + i, Name: name,
		Default: GpioPinMode[mode]}
	Pins[name] = p
	if p.IsExported() {
		return
	}
	return p.Export()
}
