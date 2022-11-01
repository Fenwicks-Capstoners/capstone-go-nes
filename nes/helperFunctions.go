package nes

func getBit(bit int, value uint8) bool {
	return (0x1<<bit)&value > 0
}
