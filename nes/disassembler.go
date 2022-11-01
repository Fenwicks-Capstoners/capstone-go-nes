package nes

import (
	"fmt"
)

type opCodeAndAddrMode struct {
	name     string                                    //opcode mnemonics
	addrMode func(addr uint16, bus *BUS) (int, string) //address mode function returns the size and menomic of the instruction
}

// addressing mode instruction sizes
// implied: 1 byte
// accumulator: 1 byte
// immediate: 2 bytes
// relative: 2 bytes
// absolute: 3 bytes
// zero-page: 2 bytes
// indirect: 3 bytes
// absolute-indexed: 3 bytes
// zero-page indexed: 2 bytes
// indexed indirect (X): 2 bytes
// indirect indexed (Y): 2 bytes

var opcodeNameTable = [256]opCodeAndAddrMode{
	{"BRK", implied}, {"ORA", indexIndirect}, {"XXX", implied}, {"XXX", indexIndirect}, {"XXX", zeroPage}, {"ORA", zeroPage}, {"ASL", zeroPage}, {"XXX", zeroPage}, {"PHP", implied}, {"ORA", immediate}, {"ASL", accumulator}, {"XXX", immediate}, {"XXX", absolute}, {"ORA", absolute}, {"ASL", absolute}, {"XXX", absolute},
	{"BPL", relative}, {"ORA", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"ORA", zeroPageX}, {"ASL", zeroPageX}, {"XXX", zeroPageX}, {"CLC", implied}, {"ORA", absoluteY}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", absoluteX}, {"ORA", absoluteX}, {"ASL", absoluteX}, {"XXX", absoluteX},
	{"JSR", absolute}, {"AND", indexIndirect}, {"XXX", implied}, {"XXX", indexIndirect}, {"BIT", zeroPage}, {"AND", zeroPage}, {"ROL", zeroPage}, {"XXX", zeroPage}, {"PLP", implied}, {"AND", immediate}, {"ROL", accumulator}, {"XXX", immediate}, {"BIT", absolute}, {"AND", absolute}, {"ROL", absolute}, {"XXX", absolute},
	{"BMI", relative}, {"AND", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"AND", zeroPageX}, {"ROL", zeroPageX}, {"XXX", zeroPageX}, {"SEC", implied}, {"AND", absoluteY}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", absoluteX}, {"AND", absoluteX}, {"ROL", absoluteX}, {"XXX", absoluteX},
	{"RTI", implied}, {"EOR", indexIndirect}, {"XXX", implied}, {"XXX", indexIndirect}, {"XXX", zeroPage}, {"EOR", zeroPage}, {"LSR", zeroPage}, {"XXX", zeroPage}, {"PHA", implied}, {"EOR", immediate}, {"LSR", accumulator}, {"XXX", immediate}, {"JMP", absolute}, {"EOR", absolute}, {"LSR", absolute}, {"XXX", absolute},
	{"BVC", relative}, {"EOR", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"EOR", zeroPageX}, {"LSR", zeroPageX}, {"XXX", zeroPageX}, {"CLI", implied}, {"EOR", absoluteY}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", absoluteX}, {"EOR", absoluteX}, {"LSR", absoluteX}, {"XXX", absoluteX},
	{"RTS", implied}, {"ADC", indexIndirect}, {"XXX", implied}, {"XXX", indexIndirect}, {"XXX", zeroPage}, {"ADC", zeroPage}, {"ROR", zeroPage}, {"XXX", zeroPage}, {"PLA", implied}, {"ADC", immediate}, {"ROR", accumulator}, {"XXX", immediate}, {"JMP", indirect}, {"ADC", absolute}, {"ROR", absolute}, {"XXX", absolute},
	{"BVS", relative}, {"ADC", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"ADC", zeroPageX}, {"ROR", zeroPageX}, {"XXX", zeroPageX}, {"SEI", implied}, {"ADC", absoluteY}, {"XXX", implied}, {"XXX", indexIndirect}, {"XXX", absoluteX}, {"ADC", absoluteX}, {"ROR", absoluteX}, {"XXX", absoluteX},
	{"XXX", immediate}, {"STA", indexIndirect}, {"XXX", immediate}, {"XXX", indexIndirect}, {"STY", zeroPage}, {"STA", zeroPage}, {"STX", zeroPage}, {"XXX", zeroPage}, {"DEY", implied}, {"XXX", immediate}, {"TXA", implied}, {"XXX", immediate}, {"STY", absolute}, {"STA", absolute}, {"STX", absolute}, {"XXX", absolute},
	{"BCC", relative}, {"STA", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"STY", zeroPageX}, {"STA", zeroPageX}, {"STX", zeroPageY}, {"XXX", zeroPageY}, {"TYA", implied}, {"STA", absoluteY}, {"TXS", implied}, {"XXX", absoluteY}, {"XXX", absoluteX}, {"STA", absoluteX}, {"XXX", absoluteY}, {"XXX", absoluteY},
	{"LDY", immediate}, {"LDA", indexIndirect}, {"LDX", immediate}, {"XXX", indexIndirect}, {"LDY", zeroPage}, {"LDA", zeroPage}, {"LDX", zeroPage}, {"XXX", zeroPage}, {"TAY", implied}, {"LDA", immediate}, {"TAX", implied}, {"XXX", immediate}, {"LDY", absolute}, {"LDA", absolute}, {"LDX", absolute}, {"XXX", absolute},
	{"BCS", relative}, {"LDA", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"LDY", zeroPageX}, {"LDA", zeroPageX}, {"LDX", zeroPageY}, {"XXX", zeroPageY}, {"CLV", implied}, {"LDA", absoluteY}, {"TSX", implied}, {"XXX", absoluteY}, {"LDY", absoluteX}, {"LDA", absoluteX}, {"LDX", absoluteY}, {"XXX", absoluteY},
	{"CPY", immediate}, {"CMP", indexIndirect}, {"XXX", immediate}, {"XXX", indexIndirect}, {"CPY", zeroPage}, {"CMP", zeroPage}, {"DEC", zeroPage}, {"XXX", zeroPage}, {"INY", implied}, {"CMP", immediate}, {"DEX", implied}, {"XXX", immediate}, {"CPY", absolute}, {"CMP", absolute}, {"DEC", absolute}, {"XXX", absolute},
	{"BNE", relative}, {"CMP", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"CMP", zeroPageX}, {"DEC", zeroPageX}, {"XXX", zeroPageX}, {"CLD", implied}, {"CMP", absoluteY}, {"XXX", implied}, {"XXX", absoluteY}, {"XXX", absoluteX}, {"CMP", absoluteX}, {"DEC", absoluteX}, {"XXX", absoluteX},
	{"CPX", immediate}, {"SBC", indexIndirect}, {"XXX", immediate}, {"XXX", indexIndirect}, {"CPX", zeroPage}, {"SBC", zeroPage}, {"INC", zeroPage}, {"XXX", zeroPage}, {"INX", implied}, {"SBC", immediate}, {"NOP", implied}, {"XXX", immediate}, {"CPX", absolute}, {"SBC", absolute}, {"INC", absolute}, {"XXX", absolute},
	{"BEQ", relative}, {"SBC", indirectIndex}, {"XXX", implied}, {"XXX", indirectIndex}, {"XXX", zeroPageX}, {"SBC", zeroPageX}, {"INC", zeroPageX}, {"XXX", zeroPageX}, {"SED", implied}, {"SBC", absoluteY}, {"XXX", implied}, {"XXX", absoluteY}, {"XXX", absoluteX}, {"SBC", absoluteX}, {"INC", absoluteX}, {"XXX", absoluteX},
}

// DisassembleInstruction takes a BUS and address and returns
// the string representation of the instruction and the size of that instruction
func DiassembleInstruction(bus *BUS, addr uint16) (string, int) {
	instr := opcodeNameTable[bus.GetCPUByte(addr)]
	size, operand := instr.addrMode(addr+1, bus)
	return fmt.Sprintf("%s %s", instr.name, operand), size
}

// Address Mode functions return the operand and the addr of the next instruction in memory
func implied(addr uint16, bus *BUS) (int, string) {
	return 1, ""
}
func indexIndirect(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("($%02X, X)", bus.GetCPUByte(addr))
}
func zeroPage(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("$%02X", bus.GetCPUByte(addr))
}
func immediate(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("#$%02X", bus.GetCPUByte(addr))
}
func accumulator(addr uint16, bus *BUS) (int, string) {
	return 1, "A"
}
func absolute(addr uint16, bus *BUS) (int, string) {
	return 3, fmt.Sprintf("$%02X%02X", bus.GetCPUByte(addr+1), bus.GetCPUByte(addr))
}
func relative(addr uint16, bus *BUS) (int, string) {
	//outputs absolute address instead of relative offset to match output of
	// a disassembler I used as a reference for correct output
	offset := uint16(bus.GetCPUByte(addr))
	if offset&0x80 > 0 {
		offset |= 0xFF00
	}
	return 2, fmt.Sprintf("$%04X", (addr+1)+offset)
}
func indirectIndex(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("($%02X), Y", bus.GetCPUByte(addr))
}
func zeroPageX(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("$%02X, X", bus.GetCPUByte(addr))
}
func zeroPageY(addr uint16, bus *BUS) (int, string) {
	return 2, fmt.Sprintf("$%02X, Y", bus.GetCPUByte(addr))
}
func absoluteX(addr uint16, bus *BUS) (int, string) {
	return 3, fmt.Sprintf("$%02X%02X, X", bus.GetCPUByte(addr+1), bus.GetCPUByte(addr))
}
func absoluteY(addr uint16, bus *BUS) (int, string) {
	return 3, fmt.Sprintf("$%02X%02X, Y", bus.GetCPUByte(addr+1), bus.GetCPUByte(addr))
}
func indirect(addr uint16, bus *BUS) (int, string) {
	return 3, fmt.Sprintf("($%02X%02X)", bus.GetCPUByte(addr+1), bus.GetCPUByte(addr))
}
