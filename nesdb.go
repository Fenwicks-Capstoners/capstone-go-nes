// Author: Max Smoot
// NESDB = NES Debugger
// CLI debugger for debugging the 6502 emulation
// supports printing registers and examining memory.
// Uses same format specifers as GDB
// format specifiers: x (hex), b(binary), i (instruction), d (decimal, default value if not specified)
// valid number formats: 0xFFFF (hex), 0b0001 (binary), 1234 (decimal, default)
// valid commands:
// set <register or address> = <hex, binary, or decimal number>
// x, prints content in memory at provided address. Either literal number address of pc for program counter
// p, prints the contents of the cpu's register: ex p x prints the x register
// cur, prints the current instruction and how many cycles remaining in the execution of the instruction
// clock, clocks the cpu
// ni, executes next instruction
// clear, clears the terminal
// quit, quits the application
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/MaxSmoot/NES_Emulator/nes"
)

var bus = nes.CreateBus()
var cpu = nes.CreateCPU(bus)

// loadBinary loads the binary specified by --binary into memory
// at the starting address specified by -load flag
func loadBinary(path string, loadAddr uint16) bool {

	file, error := os.Open(path)
	if error != nil {
		fmt.Println("Invalid Path")
		return false
	}
	buf, error := io.ReadAll(file)
	if error != nil {
		fmt.Println("Error Reading file")
		return false
	}
	if len(buf) > nes.MemorySize {
		fmt.Println("File exceeds 65KB")
		return false
	}
	fmt.Printf("%d bytes loaded\n", copy(cpu.Bus.Memory[loadAddr:], buf))
	return true
}

// boolToUint8 returns the provided bool as a uint8
// false = 0
// true = 0x1
func boolToUint8(x bool) uint8 {
	if x {
		return 1
	} else {
		return 0
	}
}

// isDigit returns if a character (byte) is a digit or not
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// get OutputFormatAndSize returns the printf format specifier (or i for instruction format), the number of bytes (or instructions)
// to print or an error
func getOutputFormatAndSize(formatSpecifier string) (string, int, error) {
	formatSpecifier = strings.ToLower(formatSpecifier)
	formatPattern := regexp.MustCompile(`^(x|p)(/(\d*)(x|d|b|i)?)?$`)
	matches := formatPattern.FindStringSubmatch(formatSpecifier)
	if len(matches) == 0 {
		return "", 0, fmt.Errorf("invalid format specifier")
	}
	format := "%d"
	switch matches[4] {
	case "x":
		format = "%02X"
	case "b":
		format = "%08b"
	case "i":
		format = "i"
	}
	numBytes := 1
	if matches[3] != "" {
		num, err := strconv.ParseInt(matches[3], 10, 32) //can't error since regex is using \d*
		if err != nil {
			return "", 0, fmt.Errorf("invalid number of bytes specified")
		}
		numBytes = int(num)

	}
	return format, int(numBytes), nil
}

// getNumberArgument returns the uint16 specified by a string or an error
// supported formats:
// 1234 - Decimal (default)
// 0x1234 - Hex
// 0b0001001000110100 - Binary
func getNumberArgument(number string) (uint16, error) {
	base := 10
	number = strings.ToLower(number)
	numPattern := regexp.MustCompile(`(?i)^(?P<format>0x|0b)?(?P<number>[0-9a-f]+)$`)
	matches := numPattern.FindStringSubmatch(number)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid number specified")
	}
	switch matches[1] {
	case "0x":
		base = 16
	case "0b":
		base = 2
	}
	num, err := strconv.ParseUint(matches[2], base, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid number specified")
	}
	return uint16(num), nil

}

// enforce8Bits returns the value as a uint8 or an error if the provided value was too big to store in a uint8
// without loss of precision
func enforce8Bits(value uint16) (uint8, error) {
	if value&0xff00 != 0x0000 {
		return 0, fmt.Errorf("%04X is too large to store in 8bits", value)
	}
	return uint8(value), nil

}

// printMemCmd prints the value in memory at address specified by args[1] using
// format specified in args[0]
// formats supported:
// x/i prints as an instruction
// x/x prints in hex
// x/d (default) prints in decimal
// x/b prints in binary
// All formats can be prepended with a number specifying how many bytes to read. EX: x/10x to read 10 bytes in hex
// When specifying x/10i, rather than reading 10 bytes, it will read 10 instructions, using the instruction size to properly read the next instruction
func printMemCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Missing required argument")
		return
	}
	if len(args) > 2 {
		fmt.Println("Too many arguments provided")
		return
	}
	format, numBytes, err := getOutputFormatAndSize(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	var address uint16
	if strings.ToLower(args[1]) == "pc" {
		address = cpu.PC
	} else {
		num, err := getNumberArgument(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		address = num
	}
	if format == "i" {
		offset := 0
		for i := 0; i < numBytes; i++ {
			instr, size := nes.DiassembleInstruction(cpu.Bus, address+uint16(offset))
			fmt.Printf("0x%04X:\t%s\n", address+uint16(offset), instr)
			offset += size
		}
		return
	}
	for i := uint16(0); i < uint16(numBytes); i++ {
		if i%8 == 0 {
			fmt.Printf("\n0x%04X:\t", address+i)
		}
		fmt.Printf(format+" ", cpu.Bus.GetByte(address+i))
	}
	fmt.Println()

}

// printCmd prints the value stored in a cpu register
// command is p</format> <register name>
// ignores number of bytes specified. EX: p/10x is the same as p/x
// this is for code reuse
func printCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Missing required argument.")
		return
	}
	if len(args) > 2 {
		fmt.Println("Too many arguments provided.")
		return
	}
	format, bytes, err := getOutputFormatAndSize(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	if format == "i" {
		fmt.Println("Cannot print as an instruction")
		return
	}
	if bytes != 1 {
		fmt.Println("The print comand only supports the default size modifier (1)")
		return
	}
	var value uint8
	switch strings.ToUpper(args[1]) {
	//currently, printing the program counter will ignore the format and always print in HEX
	case "PC":
		fmt.Printf("0x%04X\n", cpu.PC)
		return
	case "X":
		value = cpu.X
	case "Y":
		value = cpu.Y
	case "AC":
		value = cpu.AC
	case "SR":
		value = cpu.SR
	case "S":
		value = cpu.SP
	case "CF":
		value = boolToUint8(cpu.GetFlag(nes.CF))
	case "ZF":
		value = boolToUint8(cpu.GetFlag(nes.ZF))
	case "IF":
		value = boolToUint8(cpu.GetFlag(nes.IF))
	case "DF":
		value = boolToUint8(cpu.GetFlag(nes.DF))
	case "OF":
		value = boolToUint8(cpu.GetFlag(nes.OF))
	case "NF":
		value = boolToUint8(cpu.GetFlag(nes.NF))
	case "BF":
		value = boolToUint8(cpu.GetFlag(nes.BF))
	case "OP":
		fmt.Printf("0x%04X\n", cpu.OperandAddr)
		return
	default:
		fmt.Println("Can't print " + args[1])
		return
	}
	fmt.Printf("%s:\t"+format+"\n", strings.ToUpper(args[1]), value)
}

// Sets the register specified by target to value
// If the register is an 8bit register and the provided value cannot be fit into 8bits, an error is returned
// returns true, nil if the register was set
// returns false, nil if target is not a register name
// returns false, error if the target is an 8 bit register and the value cannot fit into 8 bits
func setRegister(target string, value uint16) (bool, error) {
	registerNames := map[string]bool{"x": true, "y": true, "ac": true, "sr": true, "sp": true}
	if registerNames[target] && value&0xFF00 != 0x0000 {
		return false, fmt.Errorf("provided value cannot be stored in 8bits")
	}
	switch target {
	case "pc":
		cpu.PC = value
	case "x":
		cpu.X = uint8(value)
	case "y":
		cpu.Y = uint8(value)
	case "ac":
		cpu.AC = uint8(value)
	case "sp":
		cpu.SP = uint8(value)
	case "sr":
		cpu.SR = uint8(value)
	default:
		return false, nil
	}
	return true, nil
}

// set command
// set register, flag, or memory
func setCmd(cmd string) {
	setPattern := regexp.MustCompile(`(?i)^set ((?:0x|0b)?[0-9a-f]+|pc|ac|x|y|sr|sp) (=) ((?:0x|0b)?(?:[0-9a-f]+))\s*$`)
	matches := setPattern.FindStringSubmatch(cmd)
	if len(matches) == 0 {
		fmt.Println("Invalid command format")
		return
	}
	value, err := getNumberArgument(matches[3]) //get the value (last argument)
	if err != nil {
		fmt.Println(matches[3] + " is not a valid number")
		return
	}
	isSet, err := setRegister(matches[1], value) //try to set the register
	if err != nil {
		fmt.Println(err)
		return
	}
	//if the register was successfully set we are done
	if isSet {
		return
	}
	targetAddr, err := getNumberArgument(matches[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	convertedVal, err := enforce8Bits(value)
	if err != nil {
		fmt.Println(err)
		return
	}
	cpu.Bus.SetByte(targetAddr, convertedVal)
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// uses disassembler to print the current instruction pointed to by the program counter
func printCurrentInstr() {
	instr, _ := nes.DiassembleInstruction(bus, cpu.PC)
	fmt.Printf("0x%04X:\t%s |\tCycles left executing previous instruction: %d\n", cpu.PC, instr, cpu.RemCycles)
}

func main() {
	nes.CreateCart("./roms/super_mario.nes")
	//command line flags
	addrStrPtr := flag.String("load", "0x4020", "Starting address in memory to store ROM")
	binaryPathStrPtr := flag.String("binary", "", "Path to binary to load")
	flag.Parse()
	loadAddr, err := getNumberArgument(*addrStrPtr)
	if err != nil {
		fmt.Println("Invalid load address specified")
		return
	}
	if *binaryPathStrPtr == "" {
		fmt.Println("Missing path to binary")
		fmt.Println("Usage:\n--binary=<PATH_TO_BINARY> [--load=<address to load binary>]")
		os.Exit(1)
	}

	if !loadBinary(*binaryPathStrPtr, loadAddr) {
		fmt.Println(*binaryPathStrPtr + " could not be loaded")
		os.Exit(1)
	}

	cpu.Reset()
	fmt.Println("Program Loaded.\nAwaiting Input...")
	scanner := bufio.NewScanner(os.Stdin)
	input := ""
	for {
		fmt.Print("NESDB> ")
		scanner.Scan()
		input = scanner.Text()
		tokens := strings.Fields(input)
		if len(tokens) == 0 {
			continue
		}
		if tokens[0] == "p" || strings.HasPrefix(tokens[0], "p/") {
			printCmd(tokens)
		} else if tokens[0] == "x" || strings.HasPrefix(tokens[0], "x/") {
			printMemCmd(tokens)
		} else if tokens[0] == "clear" {
			fmt.Print("\033[H\033[2J")
		} else if tokens[0] == "ni" {
			cpu.Clock()
			for cpu.RemCycles > 0 {
				cpu.Clock()
			}
			printCurrentInstr()
		} else if tokens[0] == "clock" {
			cpu.Clock()
			printCurrentInstr()
		} else if tokens[0] == "quit" {
			os.Exit(0)
		} else if tokens[0] == "cur" {
			printCurrentInstr()
		} else if tokens[0] == "run" {
			totalCycles := 0
			var prev_pc uint16 = 0xFFFF
			for {

				// instr, size := nes.DiassembleInstruction(cpu.Bus, cpu.PC)
				// var instrBytes []string
				// for i := 0; i < 3; i++ {
				// 	if i < size {
				// 		instrBytes = append(instrBytes, fmt.Sprintf("%02X", cpu.Bus.GetByte(cpu.PC+uint16(i))))
				// 	} else {
				// 		instrBytes = append(instrBytes, "  ")
				// 	}
				// }
				// oldPc := cpu.PC
				cpu.Clock()
				if cpu.PC == prev_pc {

					fmt.Printf("PC stuck on %04X\n", cpu.PC)
					fmt.Println("Total Cycles", totalCycles)
					break
				}
				prev_pc = cpu.PC
				totalCycles++
				// fmt.Printf("%04X %s  %-13s |%02X %02X %02X %02X|%1b%1b%1b%1b%1b%1b|", oldPc, strings.Join(instrBytes, " "), instr, cpu.AC, cpu.X, cpu.Y, cpu.SP, Btoi(cpu.GetFlag(nes.NF)), Btoi(cpu.GetFlag(nes.OF)), Btoi(cpu.GetFlag(nes.DF)), Btoi(cpu.GetFlag(nes.IF)), Btoi(cpu.GetFlag(nes.ZF)), Btoi(cpu.GetFlag(nes.CF)))
				// fmt.Println(cpu.RemCycles + 1)
				for cpu.RemCycles > 0 {
					totalCycles++
					cpu.Clock()
				}

			}
		} else if tokens[0] == "set" {
			setCmd(input)
		} else {
			fmt.Println("Invalid Command")
		}

	}
}
