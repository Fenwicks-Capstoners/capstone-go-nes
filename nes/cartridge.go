package nes

import (
	"fmt"
	"io"
	"os"
)

type Cartridge struct {
	CHRRomSize       int  // character rom size in bytes
	PRGRomSize       int  //program rom size in bytes
	MirrorVertically bool //if true, tiles are arranged horrizontally
	HasBatteryRam    bool // if cartridge contains battery-backed PRG Ram ($6000-7FFF)
	//and need vertical mirroring, otherwise vertical arragnged tiles with horizontal mirroring
	CHRRom              []byte
	PRGRom              []byte
	IgnoreMirrorControl bool   // if true, ignore MirrorVertically flag and provide four-screen vram
	MapperNumber        uint8  //mapper to use
	HasTrainer          bool   //if true, there is a 512 byte trainer before the PRG ROM
	mapper              Mapper //the mapper to use
}

type Mapper interface {
	CPUGetMapAddr(address uint16) uint32
	PPUGetMapAddr(address uint16) uint32
}

// CreateCart load's an INES formatted Rom into
// a Cartirdge object and returns the new object
func CreateCart(filename string) (*Cartridge, error) {
	cart := new(Cartridge)
	if err := cart.parseHeader(filename); err != nil {
		return nil, fmt.Errorf("couldn't create Cartridge: %s", err)
	}
	println("MapperNumber", cart.MapperNumber)
	println("CHRRomSize", cart.CHRRomSize)
	println("PRGRomSize", cart.PRGRomSize)
	println("Mirror Vertically", cart.MirrorVertically)
	println("ignore Mirror control bit", cart.IgnoreMirrorControl)
	println("Has Battery Ram", cart.HasBatteryRam)
	//load roms
	cart.PRGRom = make([]byte, cart.PRGRomSize)
	cart.CHRRom = make([]byte, cart.CHRRomSize)
	if err := cart.loadRoms(filename); err != nil {
		return nil, fmt.Errorf("couldn't create cartridge, %s", err)
	}
	if err := cart.loadMapper(); err != nil {
		return nil, fmt.Errorf("coudn't create cartridge, %s", err)
	}
	return cart, nil
}

func (cart *Cartridge) loadRoms(filename string) error {
	romFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("can't open file in loadRoms(), %s", err)
	}
	romBuffer, err := io.ReadAll(romFile)
	if err != nil {
		return fmt.Errorf("can't read bytes from Rom file in loadRoms(), %s", err)
	}
	offset := 16 //offset so we start reading after the header
	if cart.HasTrainer {
		offset += 512 //offset so we aren't reading the trainer
	}
	if bytesCopied := copy(cart.PRGRom, romBuffer[offset:offset+cart.PRGRomSize]); bytesCopied != cart.PRGRomSize {
		return fmt.Errorf("header specified PRG Rom size of %d but only %d bytes read", cart.PRGRomSize, bytesCopied)
	}
	if bytesCopied := copy(cart.CHRRom, romBuffer[offset+cart.PRGRomSize:offset+cart.PRGRomSize+cart.CHRRomSize]); bytesCopied != cart.CHRRomSize {
		return fmt.Errorf("header specified CHR Rom size of %d but only %d bytes read", cart.CHRRomSize, bytesCopied)
	}
	return nil
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
	cart.PRGRomSize = 16384 * int(buffer[4]) //compute PRG Rom size
	cart.CHRRomSize = 8192 * int(buffer[5])  //compute CHR Rom size
	cart.MirrorVertically = getBit(0, buffer[6])
	cart.IgnoreMirrorControl = getBit(3, buffer[6])
	cart.MapperNumber = (buffer[6] >> 4) | (buffer[7] & 0xF0)
	if getBit(0, buffer[7]) {
		return fmt.Errorf("rom is for VS Unisystem")
	}
	if !getBit(2, buffer[7]) && getBit(3, buffer[7]) {
		return fmt.Errorf("nes 2.0 roms not supported")
	}
	return nil
}

// loadMapper use's the cartridges mapper number from the header
// to attach a mapper object to cartridge.MemMapper
func (cart *Cartridge) loadMapper() error {
	switch cart.MapperNumber {
	case 0:
		cart.mapper = CreateMapper_0(cart.PRGRomSize == 16*1024)
	default:
		return fmt.Errorf("unsupported mapper: %d", cart.MapperNumber)
	}
	return nil
}

func (cart *Cartridge) GetCPUByte(addr uint16) uint8 {
	mapped_addr := cart.mapper.CPUGetMapAddr(addr)
	return cart.PRGRom[mapped_addr]
}
func (cart *Cartridge) SetCPUByte(addr uint16, value uint8) {
	mapped_addr := cart.mapper.CPUGetMapAddr(addr)
	fmt.Printf("Addr: 0x%04X Mapped Addr: 0x%08X\n", addr, mapped_addr)
	cart.PRGRom[mapped_addr] = value
}
func (cart *Cartridge) GetPPUByte(addr uint16) uint8 {
	mapped_addr := cart.mapper.CPUGetMapAddr(addr)
	return cart.CHRRom[mapped_addr]
}
func (cart *Cartridge) SetPPUByte(addr uint16, value uint8) {
	mapped_addr := cart.mapper.CPUGetMapAddr(addr)
	cart.CHRRom[mapped_addr] = value
}
