package nes

type Mapper_0 struct {
	mirrorPRGRom bool
}

func CreateMapper_0(mirrorPRGRom bool) *Mapper_0 {
	mapper := new(Mapper_0)
	mapper.mirrorPRGRom = mirrorPRGRom
	return mapper
}

func (mapper *Mapper_0) CPUGetMapAddr(addr uint16) uint16 {
	if mapper.mirrorPRGRom {
		return (addr - 0x8000) & 0x3FFF //mirror first 16kb
	}
	return addr - 0x8000 //not mirrored
}
func (mapper *Mapper_0) PPUGetMapAddr(addr uint16) uint16 {
	if mapper.mirrorPRGRom {
		return (addr - 0x8000) & 0x3FFF //mirror first 16kb
	}
	return addr - 0x8000 //not mirrored
}
