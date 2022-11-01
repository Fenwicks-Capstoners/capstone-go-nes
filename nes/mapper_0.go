package nes

type Mapper_0 struct {
	mirrorPRGRom bool
}

func CreateMapper_0(mirrorPRGRom bool) *Mapper_0 {
	mapper := new(Mapper_0)
	mapper.mirrorPRGRom = mirrorPRGRom
	return mapper
}

func (mapper *Mapper_0) CPUGetMapAddr(addr uint16) uint32 {
	if mapper.mirrorPRGRom {
		return (uint32(addr) - 0x00008000) & 0x00003FFF //mirror first 16kb
	}
	return uint32(addr) - 0x00008000 //not mirrored
}
func (mapper *Mapper_0) PPUGetMapAddr(addr uint16) uint32 {
	if mapper.mirrorPRGRom {
		return (uint32(addr) - 0x8000) & 0x3FFF //mirror first 16kb
	}
	return uint32(addr) - 0x8000 //not mirrored
}
