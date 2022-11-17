package nes

import "fmt"

const MemorySize = 2048 //2 KB

type NesSystem struct {
	Memory []uint8    // 2 kilobyte internal ram
	Cart   *Cartridge //cartridge
	CPU    *CPU
	PPU    *PPU
}

func CreateNES(romPath string) (*NesSystem, error) {
	bus := new(NesSystem)
	cart, err := CreateCart(romPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't create bus, %s", err)
	}
	bus.Cart = cart
	bus.CPU = CreateCPU(bus)
	bus.Memory = make([]uint8, MemorySize) //initalize ram
	bus.PPU = CreatePPU(bus)
	return bus, nil
}

func (bus *NesSystem) ClockSystem() {
	//ppu clocked 3x as often as CPU

	bus.CPU.Clock()
	bus.PPU.ClockPPU()
	bus.PPU.ClockPPU()
	bus.PPU.ClockPPU()
}

func (bus *NesSystem) GetCPUByte(addr uint16) uint8 {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		return bus.Memory[addr&0x07FF] //same as % 0x800. x % 2^n == x & (2^n - 1)
	}
	//NES PPU Registers
	if addr <= 0x3FFF {
		// TODO
		return bus.PPU.Registers[(addr&0x2007)-0x2000]
	}
	//NES APU and I/O registers
	if addr <= 0x4017 {
		// TODO
		return 0
	}
	//APU and I/O functionality that is normally disabled
	if addr <= 0x401f {
		// TODO
		return 0
	}
	//cartridge space
	if addr <= 0xFFFF {
		// TODO
		return bus.Cart.GetCPUByte(addr)
	}
	panic("Unsporrted Address")
}

func (bus *NesSystem) SetCPUByte(addr uint16, value uint8) {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		bus.Memory[addr&0x07FF] = value //same as % 0x800. x % 2^n == x & (2^n - 1)
	}
	//NES PPU Registers
	if addr <= 0x3FFF {
		// TODO
		bus.PPU.Registers[(addr&0x2007)-0x2000] = value
		return
	}
	//NES APU and I/O registers
	if addr <= 0x4017 {
		// TODO
		return
	}
	//APU and I/O functionality that is normally disabled
	if addr <= 0x401f {
		// TODO
		return
	}
	//cartridge space
	if addr <= 0xFFFF {

		// TODO
		bus.Cart.SetCPUByte(addr, value)
		return
	}
	panic("Unsporrted Address")
}
