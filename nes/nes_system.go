package nes

const MemorySize = 65536

type NesSystem struct {
	Memory []uint8    // 2 kilobyte internal ram
	Cart   *Cartridge //cartridge
	CPU    *CPU
}

func CreateBus(romPath string) (*NesSystem, error) {
	bus := new(NesSystem)
	// cart, err := CreateCart(romPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("couldn't create bus, %s", err)
	// }
	// bus.Cart = cart
	bus.CPU = CreateCPU(bus)
	bus.Memory = make([]uint8, MemorySize) //initalize ram
	return bus, nil
}

func (bus *NesSystem) GetCPUByte(addr uint16) uint8 {
	//internal RAM
	return bus.Memory[addr]
}

func (bus *NesSystem) SetCPUByte(addr uint16, value uint8) {
	//internal RAM
	bus.Memory[addr] = value
}
