// Author: Max Smoot
// NESDB = NES Debugger
// CLI debugger for debugging the 6502 emulation
// supports printing registers and examining memory.
// Uses same format specifers as GDB
// format specifiers: x (hex), b(binary), i (instruction), d (decimal, default value if not specified)
// valid number formats: 0xFFFF (hex), 0b0001 (binary), 1234 (decimal, default)
// valid commands:
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

// loads binary specified by path into memory
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
	println("load", loadAddr)
	copy(cpu.Bus.Memory[loadAddr:], buf)
	// copy(cpu.Bus.Memory, buf)

	return true
}

// Converts a bool to a uint8
func boolToUint8(x bool) uint8 {
	if x {
		return 1
	} else {
		return 0
	}
}

// returns if a character (byte) is a digit or not
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// Returns the printf format specifier (or i for an instruction), the number of bytes (or instructions) to print or an error
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

// Returns the uint16 specified by a string or an error
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

// Prints the value in memory at address specified by args[1] using
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

// Prints cpu register by name
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

// set command
// set register, flag, or memory
func setCmd(args []string) {
	target := strings.ToLower(args[1])
	value, err := getNumberArgument(args[3])
	if err != nil {
		fmt.Println("Invalid value")
		return
	}
	switch target {
	case "pc":
		cpu.PC = value
		return
	case "x":
		cpu.X = uint8(value)
		return
	case "y":
		cpu.Y = uint8(value)
		return
	case "ac":
		cpu.AC = uint8(value)
		return
	case "sr":
		cpu.SR = uint8(value)
		return
	case "sp":
		cpu.SP = uint8(value)
		return

	}
	targetAddr, err := getNumberArgument(args[1])
	if err != nil {
		fmt.Println("Invalid target")
		return
	}
	cpu.Bus.SetByte(targetAddr, uint8(value))
}

// uses disassembler to print the current instruction pointed to by the program counter
func printCurrentInstr() {
	instr, _ := nes.DiassembleInstruction(bus, cpu.PC)
	fmt.Printf("0x%04X:\t%s |\tCycles left executing previous instruction: %d\n", cpu.PC, instr, cpu.RemCycles)
}

func main() {

	//command line flags
	addrStrPtr := flag.String("load", "0x4020", "Starting address in memory to store ROM")
	binaryPathStrPtr := flag.String("binary", "", "Path to binary to load")
	flag.Parse()
	loadAddr, err := getNumberArgument(*addrStrPtr)
	if err != nil {
		fmt.Println("Invalid load address specified")
		return
	}

	//if the user didn't provide a binary file as a command line argument
	if *binaryPathStrPtr == "" {
		binaryLoaded := false
		for !binaryLoaded {
			fmt.Println("Enter path to binary:")
			var path string
			fmt.Scanln(&path)
			binaryLoaded = loadBinary(path, loadAddr)
		}
	} else {
		if !loadBinary(*binaryPathStrPtr, loadAddr) {
			fmt.Println(*binaryPathStrPtr + " could not be loaded")
			os.Exit(1)
		}
	}

	cpu.Reset()
	// cpu.PC = 0x400
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
			for i := 0; i < 1000; i++ {
				cpu.Clock()
			}
		} else if tokens[0] == "set" {
			setCmd(tokens)
		} else {
			fmt.Println("Invalid Command")
		}

	}
}
