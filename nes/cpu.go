package nes

type instructionAndAddrMode struct {
	instr    func()
	addrMode func() bool //updates operand and returns true if there is the possibility
	// of an extra cycle due to a page boundary crossing
	//(carry bit in lower byte addition for indexed addressing modes)
}
type CPU struct {
	Bus              *BUS
	A                uint8                       //accumulator register
	X                uint8                       //index register
	Y                uint8                       //index register
	S                uint8                       //stack pointer
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
		{a.brk, a.implied}, {a.ora, a.indexIndirect}, {a.xxx, a.implied}, {a.xxx, a.indexIndirect}, {a.xxx, a.zeroPage}, {a.ora, a.zeroPage}, {a.asl, a.zeroPage}, {a.xxx, a.zeroPage}, {a.php, a.implied}, {a.ora, a.immediate}, {a.aslA, a.accumulator}, {a.xxx, a.immediate}, {a.xxx, a.absolute}, {a.ora, a.absolute}, {a.asl, a.absolute}, {a.xxx, a.absolute},
		{a.bpl, a.relative}, {a.ora, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.ora, a.zeroPageX}, {a.asl, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.clc, a.implied}, {a.ora, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.absoluteX}, {a.ora, a.absoluteX}, {a.asl, a.absoluteX}, {a.xxx, a.absoluteX},
		{a.jsr, a.absolute}, {a.and, a.indexIndirect}, {a.xxx, a.implied}, {a.xxx, a.indexIndirect}, {a.bit, a.zeroPage}, {a.and, a.zeroPage}, {a.rol, a.zeroPage}, {a.xxx, a.zeroPage}, {a.plp, a.implied}, {a.and, a.immediate}, {a.rolA, a.accumulator}, {a.xxx, a.immediate}, {a.bit, a.absolute}, {a.and, a.absolute}, {a.rol, a.absolute}, {a.xxx, a.absolute},
		{a.bmi, a.relative}, {a.and, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.and, a.zeroPageX}, {a.rol, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.sec, a.implied}, {a.and, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.absoluteX}, {a.and, a.absoluteX}, {a.rol, a.absoluteX}, {a.xxx, a.absoluteX},
		{a.rti, a.implied}, {a.eor, a.indexIndirect}, {a.xxx, a.implied}, {a.xxx, a.indexIndirect}, {a.xxx, a.zeroPage}, {a.eor, a.zeroPage}, {a.lsr, a.zeroPage}, {a.xxx, a.zeroPage}, {a.pha, a.implied}, {a.eor, a.immediate}, {a.lsrA, a.accumulator}, {a.xxx, a.immediate}, {a.jmp, a.absolute}, {a.eor, a.absolute}, {a.lsr, a.absolute}, {a.xxx, a.absolute},
		{a.bvc, a.relative}, {a.eor, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.eor, a.zeroPageX}, {a.lsr, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.cli, a.implied}, {a.eor, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.absoluteX}, {a.eor, a.absoluteX}, {a.lsr, a.absoluteX}, {a.xxx, a.absoluteX},
		{a.rts, a.implied}, {a.adc, a.indexIndirect}, {a.xxx, a.implied}, {a.xxx, a.indexIndirect}, {a.xxx, a.zeroPage}, {a.adc, a.zeroPage}, {a.ror, a.zeroPage}, {a.xxx, a.zeroPage}, {a.pla, a.implied}, {a.adc, a.immediate}, {a.rorA, a.accumulator}, {a.xxx, a.immediate}, {a.jmp, a.indirect}, {a.adc, a.absolute}, {a.ror, a.absolute}, {a.xxx, a.absolute},
		{a.bvs, a.relative}, {a.adc, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.adc, a.zeroPageX}, {a.ror, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.sei, a.implied}, {a.adc, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.indexIndirect}, {a.xxx, a.absoluteX}, {a.adc, a.absoluteX}, {a.ror, a.absoluteX}, {a.xxx, a.absoluteX},
		{a.xxx, a.immediate}, {a.sta, a.indexIndirect}, {a.xxx, a.immediate}, {a.xxx, a.indexIndirect}, {a.sty, a.zeroPage}, {a.sta, a.zeroPage}, {a.stx, a.zeroPage}, {a.xxx, a.zeroPage}, {a.dey, a.implied}, {a.xxx, a.immediate}, {a.txa, a.implied}, {a.xxx, a.immediate}, {a.sty, a.absolute}, {a.sta, a.absolute}, {a.stx, a.absolute}, {a.xxx, a.absolute},
		{a.bcc, a.relative}, {a.sta, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.sty, a.zeroPageX}, {a.sta, a.zeroPageX}, {a.stx, a.zeroPageY}, {a.xxx, a.zeroPageY}, {a.tya, a.implied}, {a.sta, a.absoluteY}, {a.txs, a.implied}, {a.xxx, a.absoluteY}, {a.xxx, a.absoluteX}, {a.sta, a.absoluteX}, {a.xxx, a.absoluteY}, {a.xxx, a.absoluteY},
		{a.ldy, a.immediate}, {a.lda, a.indexIndirect}, {a.ldx, a.immediate}, {a.xxx, a.indexIndirect}, {a.ldy, a.zeroPage}, {a.lda, a.zeroPage}, {a.ldx, a.zeroPage}, {a.xxx, a.zeroPage}, {a.tay, a.implied}, {a.lda, a.immediate}, {a.tax, a.implied}, {a.xxx, a.immediate}, {a.ldy, a.absolute}, {a.lda, a.absolute}, {a.ldx, a.absolute}, {a.xxx, a.absolute},
		{a.bcs, a.relative}, {a.lda, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.ldy, a.zeroPageX}, {a.lda, a.zeroPageX}, {a.ldx, a.zeroPageY}, {a.xxx, a.zeroPageY}, {a.clv, a.implied}, {a.lda, a.absoluteY}, {a.tsx, a.implied}, {a.xxx, a.absoluteY}, {a.ldy, a.absoluteX}, {a.lda, a.absoluteX}, {a.ldx, a.absoluteY}, {a.xxx, a.absoluteY},
		{a.cpy, a.immediate}, {a.cmp, a.indexIndirect}, {a.xxx, a.immediate}, {a.xxx, a.indexIndirect}, {a.cpy, a.zeroPage}, {a.cmp, a.zeroPage}, {a.dec, a.zeroPage}, {a.xxx, a.zeroPage}, {a.iny, a.implied}, {a.cmp, a.immediate}, {a.dex, a.implied}, {a.xxx, a.immediate}, {a.cpy, a.absolute}, {a.cmp, a.absolute}, {a.dec, a.absolute}, {a.xxx, a.absolute},
		{a.bne, a.relative}, {a.cmp, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.cmp, a.zeroPageX}, {a.dec, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.cld, a.implied}, {a.cmp, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.absoluteY}, {a.xxx, a.absoluteX}, {a.cmp, a.absoluteX}, {a.dec, a.absoluteX}, {a.xxx, a.absoluteX},
		{a.cpx, a.immediate}, {a.sbc, a.indexIndirect}, {a.xxx, a.immediate}, {a.xxx, a.indexIndirect}, {a.cpx, a.zeroPage}, {a.sbc, a.zeroPage}, {a.inc, a.zeroPage}, {a.xxx, a.zeroPage}, {a.inx, a.implied}, {a.sbc, a.immediate}, {a.nop, a.implied}, {a.xxx, a.immediate}, {a.cpx, a.absolute}, {a.sbc, a.absolute}, {a.inc, a.absolute}, {a.xxx, a.absolute},
		{a.beq, a.relative}, {a.sbc, a.indirectIndex}, {a.xxx, a.implied}, {a.xxx, a.indirectIndex}, {a.xxx, a.zeroPageX}, {a.sbc, a.zeroPageX}, {a.inc, a.zeroPageX}, {a.xxx, a.zeroPageX}, {a.sed, a.implied}, {a.sbc, a.absoluteY}, {a.xxx, a.implied}, {a.xxx, a.absoluteY}, {a.xxx, a.absoluteX}, {a.sbc, a.absoluteX}, {a.inc, a.absoluteX}, {a.xxx, a.absoluteX},
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
func (cpu *CPU) indexIndirect() bool {
	operandAddr := cpu.Bus.GetByte(cpu.PC)
	cpu.PC++
	cpu.Operand = cpu.Get2Bytes(uint16(operandAddr + cpu.X))
	return false
}
func (cpu *CPU) zeroPage() bool {
	cpu.Operand = uint16(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	return false
}
func (cpu *CPU) immediate() bool {
	cpu.Operand = uint16(cpu.Bus.GetByte(cpu.PC))
	cpu.PC++
	return false
}
func (cpu *CPU) accumulator() bool {
	//does nothing
	return false
}
func (cpu *CPU) absolute() bool {
	cpu.Operand = cpu.Get2Bytes(cpu.PC) //gets the 2 byte address operand
	cpu.PC += 2
	return false
}
func (cpu *CPU) relative() bool {
	//increment program counter to next instructin before adding the offset
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
func (cpu *CPU) indirectIndex() bool {
	indirectAddr := cpu.Bus.GetByte(cpu.PC)
	cpu.PC += 2
	absAddr := cpu.Get2Bytes(uint16(indirectAddr))
	cpu.Operand = absAddr + uint16(cpu.Y)
	return false
}
func (cpu *CPU) zeroPageX() bool {
	cpu.Operand = uint16(cpu.X + cpu.Bus.GetByte(cpu.PC)) //since both oeprands are uint8 it will drop the carry
	cpu.PC++
	return false
}
func (cpu *CPU) zeroPageY() bool {
	cpu.Operand = uint16(cpu.Y + cpu.Bus.GetByte(cpu.PC)) //since both oeprands are uint8 it will drop the carry
	cpu.PC++
	return false
}
func (cpu *CPU) absoluteX() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.Operand = absAddr + uint16(cpu.X)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.X) > 0xFF // return true if there was a carry
}
func (cpu *CPU) absoluteY() bool {
	absAddr := cpu.Get2Bytes(cpu.PC)
	cpu.Operand = absAddr + uint16(cpu.Y)
	cpu.PC += 2
	return (absAddr&0x00FF)+uint16(cpu.Y) > 0xFF // return true if there was a carry

}
func (cpu *CPU) indirect() bool {
	cpu.Operand = cpu.Get2Bytes(cpu.Get2Bytes(cpu.PC + 1))
	cpu.PC += 2
	return false
}

/*
Instruction Functions
*/

func (cpu *CPU) xxx() { //invalid opcode will treat as NOP for now

}
func (cpu *CPU) adc() {

}
func (cpu *CPU) and() {

}
func (cpu *CPU) asl() {

}
func (cpu *CPU) aslA() {

}
func (cpu *CPU) bcc() {

}
func (cpu *CPU) bcs() {

}
func (cpu *CPU) beq() {

}
func (cpu *CPU) bit() {

}
func (cpu *CPU) bmi() {

}
func (cpu *CPU) bne() {

}
func (cpu *CPU) bpl() {

}
func (cpu *CPU) brk() {

}
func (cpu *CPU) bvc() {

}
func (cpu *CPU) bvs() {

}
func (cpu *CPU) clc() {

}
func (cpu *CPU) cld() {

}
func (cpu *CPU) cli() {

}
func (cpu *CPU) clv() {

}
func (cpu *CPU) cmp() {

}
func (cpu *CPU) cpx() {

}
func (cpu *CPU) cpy() {

}
func (cpu *CPU) dec() {

}
func (cpu *CPU) dex() {

}
func (cpu *CPU) dey() {

}
func (cpu *CPU) eor() {

}
func (cpu *CPU) inc() {

}
func (cpu *CPU) inx() {

}
func (cpu *CPU) iny() {

}
func (cpu *CPU) jmp() {

}
func (cpu *CPU) jsr() {

}
func (cpu *CPU) lda() {

}
func (cpu *CPU) ldx() {

}
func (cpu *CPU) ldy() {

}
func (cpu *CPU) lsr() {

}
func (cpu *CPU) lsrA() {

}
func (cpu *CPU) nop() {

}
func (cpu *CPU) ora() {

}
func (cpu *CPU) pha() {

}
func (cpu *CPU) php() {

}
func (cpu *CPU) pla() {

}
func (cpu *CPU) plp() {

}
func (cpu *CPU) rol() {

}
func (cpu *CPU) rolA() {

}
func (cpu *CPU) ror() {

}
func (cpu *CPU) rorA() {

}
func (cpu *CPU) rti() {

}
func (cpu *CPU) rts() {

}
func (cpu *CPU) sbc() {

}
func (cpu *CPU) sec() {

}
func (cpu *CPU) sed() {

}
func (cpu *CPU) sei() {

}
func (cpu *CPU) sta() {

}
func (cpu *CPU) stx() {

}
func (cpu *CPU) sty() {

}
func (cpu *CPU) tax() {

}
func (cpu *CPU) tay() {

}
func (cpu *CPU) tsx() {

}
func (cpu *CPU) txa() {

}
func (cpu *CPU) txs() {

}
func (cpu *CPU) tya() {

}

func (cpu *CPU) Reset() {
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
		instruction.addrMode()
		//execute instruction
		instruction.instr()

	} else {
		//otherwise don't do anything and decrement cycles
		cpu.Cycles--
	}
}
