package nes

import "fmt"

const MemorySize = 65536 //65 KB

type BUS struct {
	Memory []uint8 // 2 kilobyte internal ram

}

func CreateBus() *BUS {
	bus := new(BUS)
	bus.Memory = make([]uint8, MemorySize) //initalize ram
	return bus
}

func (bus *BUS) getSlice(addr uint16) ([]uint8, error) {
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		return bus.Memory[addr%0x0800 : addr%0x0800+3], nil //handle mirroring by wrapping the addresses around 0x0800
	}
	if uint(addr)+3 >= 0xFFFF {
		return []uint8{}, fmt.Errorf("index out of range for dissasembling instruction")
	}
	return bus.Memory[addr:(addr + 3)], nil
}

func (bus *BUS) GetByte(addr uint16) uint8 {
	//internal RAM
	// if addr <= 0x1FFF {
	// 	//0x0000-0x07FF internal RAM
	// 	//0x0800 - 0x1FFF mirrored
	// 	return bus.Memory[addr%0x0800] //handle mirroring by wrapping the addresses around 0x0800
	// }

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
	bus.Memory[addr] = value

}
