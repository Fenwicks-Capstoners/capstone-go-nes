package nes

const MemorySize = 2048 //2 KB

type BUS struct {
	Memory []uint8 // 2 kilobyte internal ram

}

func CreateBus() *BUS {
	bus := new(BUS)
	bus.Memory = make([]uint8, MemorySize) //initalize ram
	return bus
}

func (bus *BUS) GetByte(addr uint16) uint8 {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		return bus.Memory[addr&0x07FF] //same as % 0x800. x % 2^n == x & (2^n - 1)
	}
	panic("Unsporrted Address")
}

func (bus *BUS) SetByte(addr uint16, value uint8) {
	//internal RAM
	if addr <= 0x1FFF {
		//0x0000-0x07FF internal RAM
		//0x0800 - 0x1FFF mirrored
		bus.Memory[addr&0x07FF] = value //handle mirroring by wrapping the addresses around 0x0800
		return
	}
	panic("Unsporrted Address")

}
