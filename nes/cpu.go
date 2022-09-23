package nes

type instructionAndAddrMode struct {
	instr    func() bool //runs instruction, returns true if instruction could possibly take an extra cycle
	addrMode func() bool //updates operand and returns true if there is the possibility
	// of an extra cycle due to a page boundary crossing
	//(carry bit in lower byte addition for indexed addressing modes)
	cycles int
}
type CPU struct {
	Bus              *BUS
	A                uint8                       //accumulator register
	X                uint8                       //index register
	Y                uint8                       //index register
	SP               uint8                       //stack pointer
	PC               uint16                      //program counter
	CF               bool                        // carry flag
	ZF               bool                        // zero flag
	IF               bool                        //interrupt disable flag
	DF               bool                        // decimal flag
	OF               bool                        // overflow flag
	NF               bool                        // negative flag
	Cycles           int                         //number of cycles left in current instruction
	Operand          uint16                      // the operand, sometimes a single byte, sometimes a 2 byte address
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
		{a.ldyImm, a.immediate, 2}, {a.lda, a.indexIndirect, 6}, {a.ldxImm, a.immediate, 2}, {a.xxx, a.indexIndirect, 4}, {a.ldy, a.zeroPage, 3}, {a.lda, a.zeroPage, 3}, {a.ldx, a.zeroPage, 3}, {a.xxx, a.zeroPage, 4}, {a.tay, a.implied, 2}, {a.ldaImm, a.immediate, 2}, {a.tax, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.ldy, a.absolute, 4}, {a.lda, a.absolute, 4}, {a.ldx, a.absolute, 4}, {a.xxx, a.absolute, 4},
		{a.bcs, a.relative, 2}, {a.lda, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.ldy, a.zeroPageX, 4}, {a.lda, a.zeroPageX, 4}, {a.ldx, a.zeroPageY, 4}, {a.xxx, a.zeroPageY, 4}, {a.clv, a.implied, 2}, {a.lda, a.absoluteY, 4}, {a.tsx, a.implied, 2}, {a.xxx, a.absoluteY, 4}, {a.ldy, a.absoluteX, 4}, {a.lda, a.absoluteX, 4}, {a.ldx, a.absoluteY, 4}, {a.xxx, a.absoluteY, 4},
		{a.cpy, a.immediate, 2}, {a.cmp, a.indexIndirect, 6}, {a.xxx, a.immediate, 4}, {a.xxx, a.indexIndirect, 4}, {a.cpy, a.zeroPage, 3}, {a.cmp, a.zeroPage, 3}, {a.dec, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.iny, a.implied, 2}, {a.cmp, a.immediate, 2}, {a.dex, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.cpy, a.absolute, 4}, {a.cmp, a.absolute, 4}, {a.dec, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.bne, a.relative, 2}, {a.cmp, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.cmp, a.zeroPageX, 4}, {a.dec, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.cld, a.implied, 2}, {a.cmp, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteX, 4}, {a.cmp, a.absoluteX, 4}, {a.dec, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
		{a.cpx, a.immediate, 2}, {a.sbc, a.indexIndirect, 6}, {a.xxx, a.immediate, 4}, {a.xxx, a.indexIndirect, 4}, {a.cpx, a.zeroPage, 3}, {a.sbc, a.zeroPage, 3}, {a.inc, a.zeroPage, 5}, {a.xxx, a.zeroPage, 4}, {a.inx, a.implied, 2}, {a.sbc, a.immediate, 2}, {a.nop, a.implied, 2}, {a.xxx, a.immediate, 4}, {a.cpx, a.absolute, 4}, {a.sbc, a.absolute, 4}, {a.inc, a.absolute, 6}, {a.xxx, a.absolute, 4},
		{a.beq, a.relative, 2}, {a.sbc, a.indirectIndex, 5}, {a.xxx, a.implied, 4}, {a.xxx, a.indirectIndex, 4}, {a.xxx, a.zeroPageX, 4}, {a.sbc, a.zeroPageX, 4}, {a.inc, a.zeroPageX, 6}, {a.xxx, a.zeroPageX, 4}, {a.sed, a.implied, 2}, {a.sbc, a.absoluteY, 4}, {a.xxx, a.implied, 4}, {a.xxx, a.absoluteY, 4}, {a.xxx, a.absoluteX, 4}, {a.sbc, a.absoluteX, 4}, {a.inc, a.absoluteX, 7}, {a.xxx, a.absoluteX, 4},
	}
}

func (cpu *CPU) GetOperand() uint16 {
	return cpu.Operand
}

// Returns 2 bytes: addr and addr + 1
// swaps bytes due to little endian encoding, returns 16 bit number
func (cpu *CPU) Get2Bytes(addr uint16) uint16 {
	lowerByte := uint16(cpu.Bus.GetByte(addr))
	upperByte := uint16(cpu.Bus.GetByte(addr + 1))
	return (upperByte << 8) | lowerByte
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
	cpu.PC++
	cpu.Operand = cpu.Get2Bytes(uint16(operandAddr + cpu.X))
	return false
}

// operand is the byte after the instruction byte (zero page address)
func (cpu *CPU) zeroPage() bool {
	cpu.Operand = uint16(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	return false
}

// operand is the byte after the instruction byte
func (cpu *CPU) immediate() bool {
	cpu.Operand = uint16(cpu.Bus.GetByte(cpu.PC))
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
	cpu.Operand = cpu.Get2Bytes(cpu.PC) //gets the 2 byte address operand
	cpu.PC += 2
	return false
}

// operand is the PC + single byte specified after instruction (signed)
func (cpu *CPU) relative() bool {
	//increment program counter to next instruction before adding the offset
	offset := int8(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	isNeg := offset < 0
	if isNeg {
		cpu.Operand = cpu.PC - uint16(-1*offset)
	} else {
		cpu.Operand = cpu.PC + uint16(offset)
	}
	return false
}

// byte following instruction byte is a zero page address. The operand becomes the 16bit address stored at
// that location + Y
// can take an extra cycle if the read crosses a page boundary
func (cpu *CPU) indirectIndex() bool {
	indirectAddr := cpu.Bus.GetByte(cpu.PC)
	cpu.PC += 2
	absAddr := cpu.Get2Bytes(uint16(indirectAddr))
	cpu.Operand = absAddr + uint16(cpu.Y)
	return (absAddr&0x00FF)+uint16(cpu.Y) > 0xFF // return true if there was a carry
}

// operand becomes immediate 1 byte value + x
func (cpu *CPU) zeroPageX() bool {
	cpu.Operand = uint16(cpu.X + cpu.Bus.GetByte(cpu.PC)) //since both operands are uint8 it will drop the carry
	cpu.PC++
	return false
}

// operand becomes immediate 1 byte value + Y
func (cpu *CPU) zeroPageY() bool {
	cpu.Operand = uint16(cpu.Y + cpu.Bus.GetByte(cpu.PC)) //since both oeprands are uint8 it will drop the carry
	cpu.PC++
	return false
}

// operand is a 16 bit immediate value + X
// can take an extra cycle if the memory read crosses a page boundary
func (cpu *CPU) absoluteX() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.Operand = absAddr + uint16(cpu.X)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.X) > 0xFF // return true if there was a carry
}

// operand is a 16 bit immediate value + Y
// can take an extra cycle if the memory read crosses a page boundary
func (cpu *CPU) absoluteY() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.Operand = absAddr + uint16(cpu.Y)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.Y) > 0xFF // return true if there was a carry

}

// only used with JMP
// operand is address stored in memory at location specified by 16 bit immediate value
func (cpu *CPU) indirect() bool {
	cpu.Operand = cpu.Get2Bytes(cpu.Get2Bytes(cpu.PC))
	cpu.PC += 2
	return false
}

/*
Instruction Functions
*/

func (cpu *CPU) xxx() bool { //invalid opcode will treat as NOP for now
	return false
}
func (cpu *CPU) adc() bool {
	return false
}
func (cpu *CPU) and() bool {
	return false

}
func (cpu *CPU) asl() bool {
	return false

}
func (cpu *CPU) aslA() bool {
	return false

}
func (cpu *CPU) bcc() bool {
	return false

}
func (cpu *CPU) bcs() bool {
	return false

}
func (cpu *CPU) beq() bool {
	return false

}
func (cpu *CPU) bit() bool {
	return false

}
func (cpu *CPU) bmi() bool {
	return false

}
func (cpu *CPU) bne() bool {
	return false

}
func (cpu *CPU) bpl() bool {
	return false

}
func (cpu *CPU) brk() bool {
	return false

}
func (cpu *CPU) bvc() bool {
	return false

}
func (cpu *CPU) bvs() bool {
	return false

}
func (cpu *CPU) clc() bool {
	return false

}
func (cpu *CPU) cld() bool {
	return false

}
func (cpu *CPU) cli() bool {
	return false

}
func (cpu *CPU) clv() bool {
	return false

}
func (cpu *CPU) cmp() bool {
	return false

}
func (cpu *CPU) cpx() bool {
	return false

}
func (cpu *CPU) cpy() bool {
	return false

}
func (cpu *CPU) dec() bool {
	return false

}
func (cpu *CPU) dex() bool {
	return false

}
func (cpu *CPU) dey() bool {
	return false

}
func (cpu *CPU) eor() bool {
	return false

}
func (cpu *CPU) inc() bool {
	return false

}
func (cpu *CPU) inx() bool {
	return false

}
func (cpu *CPU) iny() bool {
	return false

}
func (cpu *CPU) jmp() bool {
	return false

}
func (cpu *CPU) jsr() bool {
	return false

}

// load memory into Accumulator
func (cpu *CPU) lda() bool {
	cpu.A = cpu.Bus.GetByte(cpu.Operand)
	return true
}

// load immediate value into Accumulator
func (cpu *CPU) ldaImm() bool {
	cpu.A = uint8(cpu.Operand)
	return true
}

// load memory into register X
func (cpu *CPU) ldx() bool {
	cpu.X = cpu.Bus.GetByte(cpu.Operand)
	return true
}

// load immediate value into register X
func (cpu *CPU) ldxImm() bool {
	cpu.X = uint8(cpu.Operand)
	return true
}

// load memory into register Y
func (cpu *CPU) ldy() bool {
	cpu.Y = cpu.Bus.GetByte(cpu.Operand)
	return true
}

// load immediate value into register Y
func (cpu *CPU) ldyImm() bool {
	cpu.Y = uint8(cpu.Operand)
	return true
}
func (cpu *CPU) lsr() bool {
	return false

}
func (cpu *CPU) lsrA() bool {
	return false

}
func (cpu *CPU) nop() bool {
	return false

}
func (cpu *CPU) ora() bool {
	return false

}
func (cpu *CPU) pha() bool {
	return false

}
func (cpu *CPU) php() bool {
	return false

}
func (cpu *CPU) pla() bool {
	return false

}
func (cpu *CPU) plp() bool {
	return false

}
func (cpu *CPU) rol() bool {
	return false

}
func (cpu *CPU) rolA() bool {
	return false

}
func (cpu *CPU) ror() bool {
	return false

}
func (cpu *CPU) rorA() bool {
	return false

}
func (cpu *CPU) rti() bool {
	return false

}
func (cpu *CPU) rts() bool {
	return false

}
func (cpu *CPU) sbc() bool {
	return false

}
func (cpu *CPU) sec() bool {
	return false

}
func (cpu *CPU) sed() bool {
	return false

}
func (cpu *CPU) sei() bool {
	return false

}

// store accumulator in memory
func (cpu *CPU) sta() bool {
	cpu.Bus.SetByte(cpu.Operand, cpu.A)
	return false

}

// store index X in memory
func (cpu *CPU) stx() bool {
	cpu.Bus.SetByte(cpu.Operand, cpu.X)
	return false

}

// store index Y in memory
func (cpu *CPU) sty() bool {
	cpu.Bus.SetByte(cpu.Operand, cpu.Y)

	return false

}

// transfer accumulator to index x
func (cpu *CPU) tax() bool {
	cpu.X = cpu.A
	cpu.NF = (cpu.X & 0b10000000) == 1 //set negative flag
	cpu.ZF = cpu.X == 0
	return false

}

// transfer accumulator to index Y
func (cpu *CPU) tay() bool {
	cpu.Y = cpu.A
	cpu.NF = (cpu.Y & 0b10000000) == 1 //set negative flag
	cpu.ZF = cpu.Y == 0
	return false

}

// transfer stack pointer to index X
func (cpu *CPU) tsx() bool {
	cpu.X = cpu.SP
	cpu.NF = (cpu.X & 0b10000000) == 1 //set negative flag
	cpu.ZF = cpu.X == 0
	return false

}

// transfer index x to accumulator
func (cpu *CPU) txa() bool {
	cpu.A = cpu.X
	cpu.NF = (cpu.A & 0b10000000) == 1 //set negative flag
	cpu.ZF = cpu.A == 0
	return false

}

// transfer index x to stack register
func (cpu *CPU) txs() bool {
	cpu.SP = cpu.X
	return false

}

// transfer index y to accumulator
func (cpu *CPU) tya() bool {
	cpu.A = cpu.Y
	cpu.NF = (cpu.A & 0b10000000) == 1 //set negative flag
	cpu.ZF = cpu.A == 0
	return false

}

func (cpu *CPU) Reset() {
	cpu.SP = 0xFF //stack starts at 0x01FF and grows down
	cpu.PC = cpu.Get2Bytes(0xFFFC)
}

// Cycles the cpu
func (cpu *CPU) Clock() {
	if cpu.Cycles == 0 {
		//decode instruction
		instruction := cpu.instructionTable[cpu.Bus.GetByte(cpu.PC)]
		//increment program counter
		cpu.PC++
		// run address mode function to populate operand
		extraAddr := instruction.addrMode()
		//execute instruction
		extraIns := instruction.instr()
		cpu.Cycles = instruction.cycles - 1 // subtracing 1 since we executed a cycle
		if extraAddr && extraIns {
			cpu.Cycles++
		}

	} else {
		//otherwise don't do anything and decrement cycles
		cpu.Cycles--
	}
}
