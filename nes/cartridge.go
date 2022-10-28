package nes

import (
	"fmt"
	"io"
	"os"
)

type Cartridge struct {
	CHRRomSize int // character rom size in bytes
	PRGRomSize int //program rom size in bytes
	CHRRom     []byte
	PRGRom     []byte
}

// CreateCart load's an INES formatted Rom into
// a Cartirdge object and returns the new object
func CreateCart(filename string) (*Cartridge, error) {
	cart := new(Cartridge)
	if err := cart.parseHeader(filename); err != nil {
		return nil, fmt.Errorf("couldn't create Cartridge: %s", err)
	}
	return nil, fmt.Errorf("not implmented yet")
}

// parseHeader takes the path to the rom
// and populates the cartridges rom info
// from the iNES header on the rom file
func (cart *Cartridge) parseHeader(filename string) error {
	file, err := os.Open(filename) //read the iNES rom
	if err != nil {
		return fmt.Errorf("could not open %s", filename)
	}
	buffer := make([]byte, 16)
	_, error := io.ReadFull(file, buffer)
	//parse headers
	if error != nil {
		return fmt.Errorf("error reading rom into buffer")
	}
	cart.PRGRomSize = 16384 * int(buffer[4])
	cart.CHRRomSize = 16384 * int(buffer[5])
	return nil
}
