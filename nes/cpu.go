package nes

const NF = 7
const OF = 6
const BF = 4
const DF = 3
const IF = 2
const ZF = 1
const CF = 0

type instructionAndAddrMode struct {
	instr    func() bool //runs instruction, returns true if instruction could possibly take an extra cycle
	addrMode func() bool //updates operand and returns true if there is the possibility
	// of an extra cycle due to a page boundary crossing
	//(carry bit in lower byte addition for indexed addressing modes)
	cycles int
}
type CPU struct {
	Bus *BUS
	AC  uint8  //accumulator register
	X   uint8  //index register
	Y   uint8  //index register
	SR  uint8  //status register
	SP  uint8  //stack pointer
	PC  uint16 //program counter
	//helper fields
	RemCycles        int                         //cycles left in current instruction
	OperandAddr      uint16                      // the operand, sometimes a single byte, sometimes a 2 byte address
	instructionTable [256]instructionAndAddrMode //maps first instruction byte to instruction function
}

func CreateCPU(bus *BUS) *CPU {
	cpu := new(CPU)
	cpu.populateInstructionTable()
	cpu.Bus = bus
	return cpu
}
func (a *CPU) populateInstructionTable() {
	a.instructionTable = [256]instructionAndAddrMode{
		{a.brk, a.implied, 7}, {a.ora, a.indexIndirect, 6}, {a.xxx, a.implied, 4}, {a.xxx, a.indexIndirect, 4}, {a.xxx, a.zeroPage, 4}, {a.ora, a.zeroPage, 3}, {a.asl, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.php, a.implied, 3}, {a.ora, a.immediate, 2}, {a.aslA, a.accumulator, 2}, {a.xxx, a.immediate, 4}, {a.xxx, a.absolute, 4}, {a.ora, a.absolute, 4}, {a.asl, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bpl, a.relative, 2}, {a.ora, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.ora, a.zeroPageX, 4}, {a.asl, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.clc, a.implied, 2}, {a.ora, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.absoluteX, 4}, {a.ora, a.absoluteX, 4}, {a.asl, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.jsr, a.absolute, 6}, {a.and, a.indexIndirect, 6}, {a.xxx, a.implied, 4}, {a.xxx, a.indexIndirect, 4}, {a.bit, a.zeroPage, 3}, {a.and, a.zeroPage, 3}, {a.rol, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.plp, a.implied, 4}, {a.and, a.immediate, 2}, {a.rolA, a.accumulator, 2}, {a.xxx, a.immediate, 4}, {a.bit, a.absolute, 4}, {a.and, a.absolute, 4}, {a.rol, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bmi, a.relative, 2}, {a.and, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.and, a.zeroPageX, 4}, {a.rol, a.zeroPageX, 5}, {a.xxx, a.zeroPageX, 4}, {a.sec, a.implied, 2}, {a.and, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.absoluteX, 4}, {a.and, a.absoluteX, 4}, {a.rol, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.rti, a.implied, 6}, {a.eor, a.indexIndirect, 6}, {a.xxx, a.implied, 4}, {a.xxx, a.indexIndirect, 4}, {a.xxx, a.zeroPage, 4}, {a.eor, a.zeroPage, 3}, {a.lsr, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.pha, a.implied, 3}, {a.eor, a.immediate, 2}, {a.lsrA, a.accumulator, 2}, {a.xxx, a.immediate, 4}, {a.jmp, a.absolute, 3}, {a.eor, a.absolute, 4}, {a.lsr, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bvc, a.relative, 2}, {a.eor, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.eor, a.zeroPageX, 4}, {a.lsr, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.cli, a.implied, 2}, {a.eor, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.absoluteX, 4}, {a.eor, a.absoluteX, 4}, {a.lsr, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.rts, a.implied, 6}, {a.adc, a.indexIndirect, 6}, {a.xxx, a.implied, 4}, {a.xxx, a.indexIndirect, 4}, {a.xxx, a.zeroPage, 4}, {a.adc, a.zeroPage, 3}, {a.ror, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.pla, a.implied, 4}, {a.adc, a.immediate, 2}, {a.rorA, a.accumulator, 2}, {a.xxx, a.immediate, 4}, {a.jmp, a.indirect, 5}, {a.adc, a.absolute, 4}, {a.ror, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bvs, a.relative, 2}, {a.adc, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.adc, a.zeroPageX, 4}, {a.ror, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.sei, a.implied, 2}, {a.adc, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.indexIndirect, 4}, {a.xxx, a.absoluteX, 4}, {a.adc, a.absoluteX, 4}, {a.ror, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.xxx, a.immediate, 4}, {a.sta, a.indexIndirect, 6}, {a.xxx, a.immediate, 4}, {a.xxx, a.indexIndirect, 4}, {a.sty, a.zeroPage, 3}, {a.sta, a.zeroPage, 3}, {a.stx, a.zeroPage, 3}, {a.xxx, a.zeroPage, 4}, {a.dey, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.txa, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.sty, a.absolute, 4}, {a.sta, a.absolute, 4}, {a.stx, a.absolute, 4}, {a.xxx, a.absolute, 4},
		{a.bcc, a.relative, 2}, {a.sta, a.indirectIndex, 6}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.sty, a.zeroPageX, 4}, {a.sta, a.zeroPageX, 4}, {a.stx, a.zeroPageY, 4}, {a.xxx, a.zeroPageY, 4}, {a.tya, a.implied, 2}, {a.sta, a.absoluteY, 5}, {a.txs, a.implied, 2}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteX, 4}, {a.sta, a.absoluteX, 5}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteY, 4},
		{a.ldy, a.immediate, 2}, {a.lda, a.indexIndirect, 6}, {a.ldx, a.immediate, 2}, {a.xxx, a.indexIndirect, 4}, {a.ldy, a.zeroPage, 3}, {a.lda, a.zeroPage, 3}, {a.ldx, a.zeroPage, 3}, {a.xxx, a.zeroPage, 4}, {a.tay, a.implied, 2}, {a.lda, a.immediate, 2}, {a.tax, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.ldy, a.absolute, 4}, {a.lda, a.absolute, 4}, {a.ldx, a.absolute, 4}, {a.xxx, a.absolute, 4},
		{a.bcs, a.relative, 2}, {a.lda, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.ldy, a.zeroPageX, 4}, {a.lda, a.zeroPageX, 4}, {a.ldx, a.zeroPageY, 4}, {a.xxx, a.zeroPageY, 4}, {a.clv, a.implied, 2}, {a.lda, a.absoluteY, 4}, {a.tsx, a.implied, 2}, {a.xxx, a.absoluteY, 4}, {a.ldy, a.absoluteX, 4}, {a.lda, a.absoluteX, 4}, {a.ldx, a.absoluteY, 4}, {a.xxx, a.absoluteY, 4},
		{a.cpy, a.immediate, 2}, {a.cmp, a.indexIndirect, 6}, {a.xxx, a.immediate, 4}, {a.xxx, a.indexIndirect, 4}, {a.cpy, a.zeroPage, 3}, {a.cmp, a.zeroPage, 3}, {a.dec, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.iny, a.implied, 2}, {a.cmp, a.immediate, 2}, {a.dex, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.cpy, a.absolute, 4}, {a.cmp, a.absolute, 4}, {a.dec, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bne, a.relative, 2}, {a.cmp, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.cmp, a.zeroPageX, 4}, {a.dec, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.cld, a.implied, 2}, {a.cmp, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteX, 4}, {a.cmp, a.absoluteX, 4}, {a.dec, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.cpx, a.immediate, 2}, {a.sbc, a.indexIndirect, 6}, {a.xxx, a.immediate, 4}, {a.xxx, a.indexIndirect, 4}, {a.cpx, a.zeroPage, 3}, {a.sbc, a.zeroPage, 3}, {a.inc, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.inx, a.implied, 2}, {a.sbc, a.immediate, 2}, {a.nop, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.cpx, a.absolute, 4}, {a.sbc, a.absolute, 4}, {a.inc, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.beq, a.relative, 2}, {a.sbc, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.sbc, a.zeroPageX, 4}, {a.inc, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.sed, a.implied, 2}, {a.sbc, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteX, 4}, {a.sbc, a.absoluteX, 4}, {a.inc, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
	}
}

// Returns 2 bytes: addr and addr + 1
// swaps bytes due to little endian encoding, returns 16 bit number
func (cpu *CPU) Get2Bytes(addr uint16) uint16 {
	lowerByte := uint16(cpu.Bus.GetByte(addr))
	upperByte := uint16(cpu.Bus.GetByte(addr + 1))
	return (upperByte << 8) | lowerByte
}

// sets NF and ZF based on provided uint8
func (cpu *CPU) setNZFlags(register uint8) {
	cpu.setFlag(ZF, register == 0)
	cpu.setFlag(NF, register&0x80 > 0)
}

// Dealing with Flags
func (cpu *CPU) GetFlag(flag uint8) bool {
	return cpu.SR&(0x1<<flag) > 0
}

func (cpu *CPU) setFlag(flag uint8, value bool) {
	if value {
		cpu.SR |= (0x1 << flag) // bitwise or target bit with 1 to set
	} else {
		cpu.SR &^= (0x1 << flag) //bitwise and everything with 1 except the target bit (hence the and not equals &^=)
	}
}
func setBit(number *uint8, bit uint8, value bool) {
	if value {
		*number |= (0x1 << bit) // bitwise or target bit with 1 to set
	} else {
		*number &^= (0x1 << bit) //bitwise and everything with 1 except the target bit (hence the and not equals &^=)
	}
}

// pushes to stack
func (cpu *CPU) pushStack(val uint8) {
	cpu.Bus.SetByte(0x100+uint16(cpu.SP), val)
	cpu.SP--
}

// pops from the stack
func (cpu *CPU) popStack() uint8 {
	cpu.SP++
	return cpu.Bus.GetByte(0x100 + uint16(cpu.SP))
}

/*
Addressing Modes
*/
func (cpu *CPU) implied() bool {
	return false
	//does nothing
}

// add x to immediate zero page address then read the address stored at that location in memory
func (cpu *CPU) indexIndirect() bool {
	operandAddr := cpu.Bus.GetByte(cpu.PC)
	cpu.OperandAddr = cpu.Get2Bytes(uint16(operandAddr + cpu.X))
	cpu.PC++
	return false
}

// operand is the byte after the instruction byte (zero page address)
func (cpu *CPU) zeroPage() bool {
	cpu.OperandAddr = uint16(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	return false
}

// operand is the byte after the instruction byte
func (cpu *CPU) immediate() bool {
	cpu.OperandAddr = cpu.PC
	cpu.PC++
	return false
}

// does nothing since accumulator is a register
func (cpu *CPU) accumulator() bool {
	//does nothing
	return false
}

// operand is two bytes following instruction byte
func (cpu *CPU) absolute() bool {
	cpu.OperandAddr = cpu.Get2Bytes(cpu.PC) //gets the 2 byte address operand
	cpu.PC += 2
	return false
}

// operand is the PC + single byte specified after instruction (signed)
func (cpu *CPU) relative() bool {
	offset := int8(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	isNeg := offset < 0
	if isNeg {
		cpu.OperandAddr = cpu.PC - uint16(-1*offset)
	} else {
		cpu.OperandAddr = cpu.PC + uint16(offset)
	}
	return false
}

// byte following instruction byte is a zero page address. The operand becomes the 16bit address stored at
// that location + Y
// can take an extra cycle if the read crosses a page boundary
func (cpu *CPU) indirectIndex() bool {
	indirectAddr := cpu.Bus.GetByte(cpu.PC)
	absAddr := cpu.Get2Bytes(uint16(indirectAddr))
	cpu.OperandAddr = absAddr + uint16(cpu.Y)
	cpu.PC++
	return (absAddr&0x00FF)+uint16(cpu.Y) > 0xFF // return true if there was a carry
}

// operand becomes immediate 1 byte value + x
func (cpu *CPU) zeroPageX() bool {
	cpu.OperandAddr = uint16(cpu.X + cpu.Bus.GetByte(cpu.PC)) //since both operands are uint8 it will drop the carry
	cpu.PC++
	return false
}

// operand becomes immediate 1 byte value + Y
func (cpu *CPU) zeroPageY() bool {
	cpu.OperandAddr = uint16(cpu.Y + cpu.Bus.GetByte(cpu.PC)) //since both oeprands are uint8 it will drop the carry
	cpu.PC++
	return false
}

// operand is a 16 bit immediate value + X
// can take an extra cycle if the memory read crosses a page boundary
func (cpu *CPU) absoluteX() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.OperandAddr = absAddr + uint16(cpu.X)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.X) > 0xFF // return true if there was a carry
}

// operand is a 16 bit immediate value + Y
// can take an extra cycle if the memory read crosses a page boundary
func (cpu *CPU) absoluteY() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.OperandAddr = absAddr + uint16(cpu.Y)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.Y) > 0xFF // return true if there was a carry

}

// only used with JMP
// operand is address stored in memory at location specified by 16 bit immediate value
// there is a bug where of a page boundary is crossed when reading the address
// JMP ($C0FF) should fetch the low byte from 0xC0FF and the high byte from 0xC100
// however, instead of this behavior, the low byte is incremented (overflowing and becoming 00) and
// the high byte isn't also incremented so the high byte would be read from 0xC000
// https://everything2.com/title/6502+indirect+JMP+bug
func (cpu *CPU) indirect() bool {
	lowByteAddr := cpu.Get2Bytes(cpu.PC)
	highByteAddr := lowByteAddr + 1   //increment for high byte address
	if lowByteAddr&0x00FF == 0x00FF { //if a page boundary needs to be crossed when incrementing
		highByteAddr = lowByteAddr & 0xFF00 //simulate the low byte of the target addr overflowing and the upper byte not adding the carry
	}
	cpu.OperandAddr = uint16(cpu.Bus.GetByte(lowByteAddr)) | (uint16(cpu.Bus.GetByte(highByteAddr)) << 8) //combine low and high byte
	cpu.PC += 2
	return false
}

/*
Instruction Functions
*/

func (cpu *CPU) xxx() bool { //invalid opcode will treat as NOP for now
	return false
}

// add with carry
// since SBC utilizes this functionality, it is in a function that takes a literal value
func (cpu *CPU) adcValue(value uint8) bool {
	oldAc := cpu.AC
	cpu.AC += value //add accumulator and memory
	//add the carry flag
	if cpu.GetFlag(CF) {
		cpu.AC++
	}
	cpu.setNZFlags(cpu.AC)
	cpu.setFlag(CF, cpu.AC < oldAc || cpu.AC < value)                //if the final answer is smaller than either of the operands, there was a carry
	oldAcSign := (oldAc & 0x80) >> 7                                 //get the sign bit of operand 1
	valSign := (value & 0x80) >> 7                                   //get the sign bit of operand 2
	AcSign := (cpu.AC & 0x80) >> 7                                   //get the sign bit of the sum
	cpu.setFlag(OF, ^((oldAcSign^valSign)|^(oldAcSign^AcSign)) == 1) //set overflow flag
	return false
}

// add memory to accumulator with carry
// NOTE: Ignoring Decimal Mode since the NES doesn't support it
func (cpu *CPU) adc() bool {
	//call the core function with the memory value
	return cpu.adcValue(cpu.Bus.GetByte(cpu.OperandAddr))

}

// sets accumulator to accumulator & value
func (cpu *CPU) and() bool {
	cpu.AC = cpu.AC & cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.AC)
	return false
}

// shift left one bit (memory)
func (cpu *CPU) asl() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setFlag(CF, value&0x80 > 0) //set CF to bit 7 since it is the bit being shifted out
	value <<= 1
	cpu.Bus.SetByte(cpu.OperandAddr, value)
	cpu.setNZFlags(value)
	return false

}

// shit left one bit (accumulator)
func (cpu *CPU) aslA() bool {
	cpu.setFlag(CF, cpu.AC&0x80 > 0) //set CF to bit 7 since it is the bit being shifted out
	cpu.AC <<= 1
	cpu.Bus.SetByte(cpu.OperandAddr, cpu.AC)
	cpu.setNZFlags(cpu.AC)
	return false

}

// branches if condition is true
func (cpu *CPU) branch(condition bool) bool {
	if condition {
		cpu.RemCycles++                            //add 1 cycle if branch on same page, otherwise add 2. In either case we increment cycles at least once
		if cpu.PC&0xFF00 != cpu.OperandAddr&0xFF { //if page boundary is crossed add the second additional cycle
			cpu.RemCycles++
		}
		cpu.PC = cpu.OperandAddr //branch PC
	}
	return false

}

// branch on carry clear
// CF = False
func (cpu *CPU) bcc() bool {
	return cpu.branch(!cpu.GetFlag(CF))
}

// branch on carry set
// CF = True
func (cpu *CPU) bcs() bool {
	return cpu.branch(cpu.GetFlag(CF))
}

// branch on result zero
// ZF = true
func (cpu *CPU) beq() bool {
	return cpu.branch(cpu.GetFlag(ZF))

}
func (cpu *CPU) bit() bool {
	return false

}

// branch on result minus
// NF = true
func (cpu *CPU) bmi() bool {
	return cpu.branch(cpu.GetFlag(NF))

}

// branch on result not zero
// ZF = False
func (cpu *CPU) bne() bool {
	return cpu.branch(!cpu.GetFlag(ZF))

}

// branch on result plus
// NF = False
func (cpu *CPU) bpl() bool {
	return cpu.branch(!cpu.GetFlag(NF))

}
func (cpu *CPU) brk() bool {
	return false

}

// branch on overflow clear
// OF = false
func (cpu *CPU) bvc() bool {
	return cpu.branch(!cpu.GetFlag(OF))

}

// branch on overflow set
// OF = true
func (cpu *CPU) bvs() bool {
	return cpu.branch(cpu.GetFlag(OF))

}

// clear the carry flag
func (cpu *CPU) clc() bool {
	cpu.setFlag(CF, false)
	return false

}

// clear decimal flag
// NOTE: NES doesn't support decimal mode so neither will this emulator
func (cpu *CPU) cld() bool {
	cpu.setFlag(DF, false)
	return false

}

// clear interrupt flag
func (cpu *CPU) cli() bool {
	cpu.setFlag(IF, false)
	return false

}

// clear overflow flag
func (cpu *CPU) clv() bool {
	cpu.setFlag(OF, false)
	return false

}

// core behvior of all compare functions
// computes Register - value
func (cpu *CPU) compareFunc(register uint8, value uint8) bool {
	cpu.setNZFlags(register - value)
	cpu.setFlag(CF, register >= value)
	return false

}

// compare memory to accumulator
func (cpu *CPU) cmp() bool {
	return cpu.compareFunc(cpu.AC, cpu.Bus.GetByte(cpu.OperandAddr))
}

// compare memory with X
func (cpu *CPU) cpx() bool {
	return cpu.compareFunc(cpu.X, cpu.Bus.GetByte(cpu.OperandAddr))
}

// compare memory to Y
func (cpu *CPU) cpy() bool {
	return cpu.compareFunc(cpu.Y, cpu.Bus.GetByte(cpu.OperandAddr))
}

// decrement memory by 1
func (cpu *CPU) dec() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	value--
	cpu.setNZFlags(value)
	cpu.Bus.SetByte(cpu.OperandAddr, value)
	return false

}

// decrement index X by 1
func (cpu *CPU) dex() bool {
	cpu.X--
	cpu.setNZFlags(cpu.X)
	return false

}

// decrement index y by 1
func (cpu *CPU) dey() bool {
	cpu.Y--
	cpu.setNZFlags(cpu.Y)
	return false

}

// eor with value from memory
func (cpu *CPU) eor() bool {
	cpu.AC ^= cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.AC)
	return false
}

// increment memory by 1
func (cpu *CPU) inc() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	value++
	cpu.setNZFlags(value)
	cpu.Bus.SetByte(cpu.OperandAddr, value)
	return false

}

// increment index x by 1
func (cpu *CPU) inx() bool {
	cpu.X++
	cpu.setNZFlags(cpu.X)
	return false

}

// increment index y by 1
func (cpu *CPU) iny() bool {
	cpu.Y++
	cpu.setNZFlags(cpu.Y)
	return false

}

// jump
func (cpu *CPU) jmp() bool {
	cpu.PC = cpu.OperandAddr
	return false

}

// jump save return
func (cpu *CPU) jsr() bool {
	cpu.pushStack(uint8(cpu.PC >> 8)) //push high byte
	cpu.pushStack(uint8(cpu.PC) - 1)  //push low byte
	//must subtract 1 because in the 6502 the pc isn't incremented by the time we push the lower byte
	cpu.PC = cpu.OperandAddr
	return false

}

// load memory into Accumulator
func (cpu *CPU) lda() bool {
	cpu.AC = cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.AC)
	return true
}

// load memory into register X
func (cpu *CPU) ldx() bool {
	cpu.X = cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.X)
	return true
}

// load memory into register Y
func (cpu *CPU) ldy() bool {
	cpu.Y = cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.Y)
	return true
}

// logical shift right with memory
func (cpu *CPU) lsr() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setFlag(CF, value&0x1 > 0)
	newValue := value >> 1
	cpu.setFlag(NF, false)
	cpu.setFlag(ZF, newValue == 0)
	cpu.Bus.SetByte(cpu.OperandAddr, newValue)
	return false

}

// logical shift right with accumulator
func (cpu *CPU) lsrA() bool {
	cpu.setFlag(CF, cpu.AC&0x1 > 0)
	cpu.AC >>= 1
	cpu.setFlag(NF, false)
	cpu.setFlag(ZF, cpu.AC == 0)
	return false

}
func (cpu *CPU) nop() bool {
	return false

}

// ora with value in memory
func (cpu *CPU) ora() bool {
	cpu.AC |= cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setNZFlags(cpu.AC)
	return false
}

// push accumulator to stack
func (cpu *CPU) pha() bool {
	cpu.pushStack(cpu.AC)
	return false

}

// push processor status to stack
func (cpu *CPU) php() bool {
	cpu.pushStack(cpu.SR | 0b00110000) //BRK and bit 5 are always 1 when not on the stack since they don't technically exist physically
	return false

}

// pull accumulator from stack
func (cpu *CPU) pla() bool {
	cpu.AC = cpu.popStack()
	cpu.setNZFlags(cpu.AC)
	return false

}

// pull processor status from stack
func (cpu *CPU) plp() bool {
	cpu.SR = cpu.popStack() | 0b00010000 //BF always true when not on the stack
	return false

}

// rotate one bit left memory
func (cpu *CPU) rol() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setFlag(CF, value&0x80 > 0) //store bit being shifted out into CF
	value <<= 1
	setBit(&value, 0, cpu.GetFlag(CF)) //perform the rotate
	cpu.Bus.SetByte(cpu.OperandAddr, value)
	cpu.setNZFlags(value)
	return false
}

// rotate one bit left accumulator
func (cpu *CPU) rolA() bool {
	cpu.setFlag(CF, cpu.AC&0x80 > 0) //store bit being shifted out into CF
	println(cpu.AC & 0x8)
	cpu.AC <<= 1
	setBit(&cpu.AC, 0, cpu.GetFlag(CF)) //perform the rotate
	cpu.setNZFlags(cpu.AC)
	return false

}

// rotate one bit right memory
func (cpu *CPU) ror() bool {
	value := cpu.Bus.GetByte(cpu.OperandAddr)
	cpu.setFlag(CF, value&0x1 > 0) //store bit being shifted out into CF
	value >>= 1
	setBit(&value, 7, cpu.GetFlag(CF)) //perform the rotate
	cpu.Bus.SetByte(cpu.OperandAddr, value)
	cpu.setNZFlags(value)
	return false
}
func (cpu *CPU) rorA() bool {
	cpu.setFlag(CF, cpu.AC&0x1 > 0) //store bit being shifted out into CF
	cpu.AC >>= 1
	setBit(&cpu.AC, 7, cpu.GetFlag(CF)) //perform the rotate
	cpu.setNZFlags(cpu.AC)
	return false

}
func (cpu *CPU) rti() bool {
	return false

}

// return from subroutine
func (cpu *CPU) rts() bool {
	pcL := uint16(cpu.popStack())
	pcH := uint16(cpu.popStack()) << 8
	cpu.PC = (pcH | pcL) + 1 //must increment since JSR saves the lower byte before the pc is incremented
	return false
}

// core functionality of sbc but uses a value parameter
// sbc will provide the value form memory
// sbcImm will provide the operand directly
// Normally 2's complement subtraction works as follows:
// a - b
// flip the bits of b and add 1 to make it negative
// then simply compute a + ^b + 1
// BUT SBC doesn't do the +1
// To achieve true subtraction, the carry flag must be set
// This is because sbc uses the same logic as ADC so it achieves ^b+1 when the carry is set
// since ADC adds the carry value
// This means to get proper subtraction, you must first set the carry flag using SEC
func (cpu *CPU) sbc() bool {
	return cpu.adcValue(cpu.Bus.GetByte(cpu.OperandAddr))

}

// set the carry flag
func (cpu *CPU) sec() bool {
	cpu.setFlag(CF, true)
	return false

}

// set decimal flag
// NOTE: NES doesnt support decimal mode so neither will this emulator
func (cpu *CPU) sed() bool {
	cpu.setFlag(DF, true)
	return false

}

// set the interrupt flag
func (cpu *CPU) sei() bool {
	cpu.setFlag(IF, true)
	return false

}

// store accumulator in memory
func (cpu *CPU) sta() bool {
	cpu.Bus.SetByte(cpu.OperandAddr, cpu.AC)
	return false

}

// store index X in memory
func (cpu *CPU) stx() bool {
	cpu.Bus.SetByte(cpu.OperandAddr, cpu.X)
	return false

}

// store index Y in memory
func (cpu *CPU) sty() bool {
	cpu.Bus.SetByte(cpu.OperandAddr, cpu.Y)
	return false

}

// transfer accumulator to index x
func (cpu *CPU) tax() bool {
	cpu.X = cpu.AC
	cpu.setNZFlags(cpu.X)

	return false

}

// transfer accumulator to index Y
func (cpu *CPU) tay() bool {
	cpu.Y = cpu.AC
	cpu.setNZFlags(cpu.Y)
	return false

}

// transfer stack pointer to index X
func (cpu *CPU) tsx() bool {
	cpu.X = cpu.SP
	cpu.setNZFlags(cpu.X)
	return false

}

// transfer index x to accumulator
func (cpu *CPU) txa() bool {
	cpu.AC = cpu.X
	cpu.setNZFlags(cpu.AC)
	return false

}

// transfer index x to stack register
func (cpu *CPU) txs() bool {
	cpu.SP = cpu.X
	return false

}

// transfer index y to accumulator
func (cpu *CPU) tya() bool {
	cpu.AC = cpu.Y
	cpu.setNZFlags(cpu.Y)
	return false

}

func (cpu *CPU) Reset() {
	cpu.SP = 0xFF                  //stack starts at 0x01FF and grows down
	cpu.SR = 0b00110000            //reset status register
	cpu.PC = cpu.Get2Bytes(0xFFFC) //retrieve program counter
}

// Cycles the cpu
func (cpu *CPU) Clock() {
	if cpu.RemCycles == 0 {
		//decode instruction
		instruction := cpu.instructionTable[cpu.Bus.GetByte(cpu.PC)]
		//increment program counter
		cpu.PC++
		// run address mode function to populate operand
		extraAddr := instruction.addrMode()
		//execute instruction
		extraIns := instruction.instr()
		cpu.RemCycles = instruction.cycles
		if extraAddr && extraIns {
			cpu.RemCycles++
		}

	}
	cpu.RemCycles--
}
