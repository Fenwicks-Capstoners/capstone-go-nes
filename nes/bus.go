package nes

import "log"

const MemorySize = 65536 //65 KB

type BUS struct {
	Memory []uint8 // 2 kilobyte internal ram

}

func CreateBus() *BUS {
	bus := new(BUS)
	bus.Memory = make([]uint8, MemorySize) //initalize ram
	return bus
}

func (bus *BUS) getSlice(addr uint16) []uint8 {
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		return bus.Memory[addr%0x0800 : addr%0x0800+3] //handle mirroring by wrapping the addresses around 0x0800
	}
	return make([]uint8, 0)
}

func (bus *BUS) GetByte(addr uint16) uint8 {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		return bus.Memory[addr%0x0800] //handle mirroring by wrapping the addresses around 0x0800
	}

	return bus.Memory[addr]
}

func (bus *BUS) SetByte(addr uint16, value uint8) {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		bus.Memory[addr%0x0800] = value //handle mirroring by wrapping the addresses around 0x0800
		return
	}
	log.Fatalf("Address %04X not supported", addr)

}
