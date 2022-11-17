package nes

//Register layout
// 0x2000 = Registers[0] = Controller
// 0x2001 =  Registers[1] = Mask
// 0x2002 = Registers[2] = Status
// 0x2003 = Registers[3] = OAM Address
// 0x2004 = Registers[4] = OAM Data
// 0x2005 = Registers[5] = Scroll
// 0x2006 = Registers[6] = Address
// 0x2007 = Registers[7] = Data
//

type AddrRegister struct {
	Addr       uint16
	isHighByte bool
}

func createAddrRegister() *AddrRegister {
	addrReg := new(AddrRegister)
	addrReg.isHighByte = true
	return addrReg
}

func (addrReg *AddrRegister) loadAddr(addrByte uint8) {
	if addrReg.isHighByte {
		addrReg.Addr &= 0x00ff
		addrReg.Addr |= (uint16(addrByte) << 8)
	} else {
		addrReg.Addr &= 0xff00
		addrReg.Addr |= uint16(addrByte)
	}
	addrReg.isHighByte = !addrReg.isHighByte
}

type PPU struct {
	Registers    []uint8 //ppu registers
	Bus          *NesSystem
	PaletteTable []uint8       // 0x3F00 - 0x3FFF
	OAM          []uint8       //object attribute memory
	VRAM         []uint8       // 0x2000 - 0x3EFF    2kB of PPU ram / Pattern Memory
	AddrRegister *AddrRegister // 16 bit address, populated by 2 Stores to 0x2006
	// EXAMPLE: LDA #$60 STA $2006 LDA #$00 STA $2006 loads 0x6000 into the PPU addr register

}

func (*PPU) ClockPPU() {

}

func CreatePPU(nes *NesSystem) *PPU {
	ppu := new(PPU)
	ppu.Bus = nes
	ppu.Registers = make([]uint8, 8)
	ppu.PaletteTable = make([]uint8, 32)
	ppu.OAM = make([]uint8, 256)
	ppu.VRAM = make([]uint8, 2048)
	ppu.AddrRegister = createAddrRegister()
	return ppu
}
