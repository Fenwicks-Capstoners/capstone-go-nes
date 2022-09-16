package nes

import (
	"fmt"
	"io"
	"log"
	"os"
)

type opCodeAndAddrMode struct {
	name     string                                 //opcode mnemonics
	addrMode func(i int, data []byte) (int, string) //address mode function
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
	{"BRK", implied}, {"ORA", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ORA", zeroPage}, {"ASL", zeroPage}, {"XXX", implied}, {"PHP", implied}, {"ORA", immediate}, {"ASL", accumulator}, {"XXX", implied}, {"XXX", implied}, {"ORA", absolute}, {"ASL", absolute}, {"XXX", immediate},
	{"BPL", relative}, {"ORA", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ORA", zeroPageX}, {"ASL", zeroPageX}, {"XXX", implied}, {"CLC", implied}, {"ORA", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ORA", absoluteX}, {"ASL", absoluteX}, {"XXX", implied},
	{"JSR", absolute}, {"AND", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"BIT", zeroPage}, {"AND", zeroPage}, {"ROL", zeroPage}, {"XXX", implied}, {"PLP", implied}, {"AND", immediate}, {"ROL", accumulator}, {"XXX", implied}, {"BIT", absolute}, {"AND", absolute}, {"ROL", absolute}, {"XXX", implied},
	{"BMI", relative}, {"AND", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"AND", zeroPageX}, {"ROL", zeroPageX}, {"XXX", implied}, {"SEC", implied}, {"AND", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"AND", absoluteX}, {"ROL", absoluteX}, {"XXX", implied},
	{"RTI", implied}, {"EOR", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"EOR", zeroPage}, {"LSR", zeroPage}, {"XXX", implied}, {"PHA", implied}, {"EOR", immediate}, {"LSR", accumulator}, {"XXX", implied}, {"JMP", absolute}, {"EOR", absolute}, {"LSR", absolute}, {"XXX", implied},
	{"BVC", relative}, {"EOR", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"EOR", zeroPageX}, {"LSR", zeroPageX}, {"XXX", implied}, {"CLI", implied}, {"EOR", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"EOR", absoluteX}, {"LSR", absoluteX}, {"XXX", implied},
	{"RTS", implied}, {"ADC", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ADC", zeroPage}, {"ROR", zeroPage}, {"XXX", implied}, {"PLA", implied}, {"ADC", immediate}, {"ROR", accumulator}, {"XXX", implied}, {"JMP", indirect}, {"ADC", absolute}, {"ROR", absolute}, {"XXX", implied},
	{"BVS", relative}, {"ADC", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ADC", zeroPageX}, {"ROR", zeroPageX}, {"XXX", implied}, {"SEI", implied}, {"ADC", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"ADC", absoluteX}, {"ROR", absoluteX}, {"XXX", implied},
	{"XXX", implied}, {"STA", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"STY", zeroPage}, {"STA", zeroPage}, {"STX", zeroPage}, {"XXX", implied}, {"DEY", implied}, {"XXX", implied}, {"TXA", implied}, {"XXX", implied}, {"STY", absolute}, {"STA", absolute}, {"STX", absolute}, {"XXX", implied},
	{"BCC", relative}, {"STA", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"STY", zeroPageX}, {"STA", zeroPageX}, {"STX", zeroPageY}, {"XXX", implied}, {"TYA", implied}, {"STA", absoluteY}, {"TXS", implied}, {"XXX", implied}, {"XXX", implied}, {"STA", absoluteX}, {"XXX", implied}, {"XXX", implied},
	{"LDY", immediate}, {"LDA", indexIndirect}, {"LDX", immediate}, {"XXX", implied}, {"LDY", zeroPage}, {"LDA", zeroPage}, {"LDX", zeroPage}, {"XXX", implied}, {"TAY", implied}, {"LDA", immediate}, {"TAX", implied}, {"XXX", implied}, {"LDY", absolute}, {"LDA", absolute}, {"LDX", absolute}, {"XXX", implied},
	{"BCS", relative}, {"LDA", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"LDY", zeroPageX}, {"LDA", zeroPageX}, {"LDX", zeroPageY}, {"XXX", implied}, {"CLV", implied}, {"LDA", absoluteY}, {"TSX", implied}, {"XXX", implied}, {"LDY", absoluteX}, {"LDA", absoluteX}, {"LDX", absoluteY}, {"XXX", implied},
	{"CPY", immediate}, {"CMP", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"CPY", zeroPage}, {"CMP", zeroPage}, {"DEC", zeroPage}, {"XXX", implied}, {"INY", implied}, {"CMP", immediate}, {"DEX", implied}, {"XXX", implied}, {"CPY", absolute}, {"CMP", absolute}, {"DEC", absolute}, {"XXX", implied},
	{"BNE", relative}, {"CMP", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"CMP", zeroPageX}, {"DEC", zeroPageX}, {"XXX", implied}, {"CLD", implied}, {"CMP", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"CMP", absoluteX}, {"DEC", absoluteX}, {"XXX", implied},
	{"CPX", immediate}, {"SBC", indexIndirect}, {"XXX", implied}, {"XXX", implied}, {"CPX", zeroPage}, {"SBC", zeroPage}, {"INC", zeroPage}, {"XXX", implied}, {"INX", implied}, {"SBC", immediate}, {"NOP", implied}, {"XXX", implied}, {"CPX", absolute}, {"SBC", absolute}, {"INC", absolute}, {"XXX", implied},
	{"BEQ", relative}, {"SBC", indirectIndex}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"SBC", zeroPageX}, {"INC", zeroPageX}, {"XXX", implied}, {"SED", implied}, {"SBC", absoluteY}, {"XXX", implied}, {"XXX", implied}, {"XXX", implied}, {"SBC", absoluteX}, {"INC", absoluteX}, {"XXX", implied},
}

func DiassembleInstruction(bus *BUS, i uint16) string {
	instMem := bus.getSlice(i)
	instr := opcodeNameTable[instMem[0]]
	_, operand := instr.addrMode(0, instMem)
	return fmt.Sprintf("0x%04X:\t%s %s", i, instr.name, operand)
}

func Disassemble(pathToBinary string) {
	file, err := os.Open(pathToBinary)
	if err != nil {
		log.Fatal("Error opening binary file", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading binary")
	}
	i := 0 //the program counter
	operand := ""
	for i < len(data) {
		// fmt.Printf("%04X\t", i)
		dataByte := data[i]
		fmt.Print(opcodeNameTable[dataByte].name + " ")
		i, operand = opcodeNameTable[dataByte].addrMode(i, data)
		fmt.Println(operand)
	}

}

// Address Mode functions return the operand and the index of the next instruction in the data buffer
func implied(i int, data []byte) (int, string) {
	return i + 1, ""
}
func indexIndirect(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("($%02X, X)", data[i+1])
}
func zeroPage(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("$%02X", data[i+1])
}
func immediate(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("#$%02X", data[i+1])
}
func accumulator(i int, data []byte) (int, string) {
	return i + 1, "A"
}
func absolute(i int, data []byte) (int, string) {
	return i + 3, fmt.Sprintf("$%02X%02X", data[i+2], data[i+1])
}
func relative(i int, data []byte) (int, string) {
	//outputs absolute address instead of relative offset to match output of
	// a disassembler I used as a reference for correct output
	return i + 2, fmt.Sprintf("$%04X", i+2+int(int8(data[i+1])))
}
func indirectIndex(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("($%02X), Y", data[i+1])
}
func zeroPageX(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("$%02X, X", data[i+1])
}
func zeroPageY(i int, data []byte) (int, string) {
	return i + 2, fmt.Sprintf("$%02X, Y", data[i+1])
}
func absoluteX(i int, data []byte) (int, string) {
	return i + 3, fmt.Sprintf("$%02X%02X, X", data[i+2], data[i+1])
}
func absoluteY(i int, data []byte) (int, string) {
	return i + 3, fmt.Sprintf("$%02X%02X, Y", data[i+2], data[i+1])
}
func indirect(i int, data []byte) (int, string) {
	return i + 3, fmt.Sprintf("$%02X%02X", data[i+2], data[i+1])
}
