// Author: Max Smoot
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
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/MaxSmoot/NES_Emulator/nes"
)

var bus = nes.CreateBus()
var cpu = nes.CreateCPU(bus)

// loads binary specified by path into memory
func loadBinary(path string) bool {

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

	copy(cpu.Bus.Memory, buf)
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

// Returns the printf format specifier (or i for an instruction), the number of bytes to print (or instructions) or an error
func getOutputFormatAndSize(formatSpecifier string) (string, int, error) {
	if len(formatSpecifier) == 1 {
		return "%d", 1, nil
	}
	formatSpecifier = strings.ToLower(formatSpecifier)
	if len(formatSpecifier) < 3 || formatSpecifier[1] != '/' {
		return "", 0, fmt.Errorf("invalid format specifier")
	}
	numberAndFormat := formatSpecifier[2:]
	numberCharacters := ""
	var formatString string
	var format byte = 'd'
	for i, char := range []byte(numberAndFormat) {
		if IsDigit(char) {
			numberCharacters += string(char)
		} else {
			format = numberAndFormat[i]
			break
		}
	}
	switch format {
	case 'x':
		formatString = "%02X"
	case 'b':
		formatString = "%08b"
	case 'd':
		formatString = "%d"
	case 'i':
		formatString = "i"
	default:
		return "", 0, fmt.Errorf("invalid format. Only x, b, or d are valid")
	}
	numBytes := 1
	if len(numberCharacters) > 0 {
		num, err := strconv.ParseInt(numberCharacters, 10, 0)
		if err != nil || num < 1 {
			return "", 0, fmt.Errorf("invalid number of bytes specified")
		}
		numBytes = int(num)
	}
	return formatString, numBytes, nil
}

// Returns the uint16 specified by a string or an error
// supported formats:
// 1234 - Decimal (default)
// 0x1234 - Hex
// 0b0001001000110100 - Binary
func getNumberArgument(number string) (uint16, error) {
	base := 10
	trimmedNumber := number
	if len(number) > 2 {
		format := strings.ToLower(number[0:2])
		if format == "0x" { //hex format
			base = 16
			trimmedNumber = number[2:]
		} else if format == "0b" { //binary format
			base = 2
			trimmedNumber = number[2:]
		} else if !IsDigit(number[0]) || !IsDigit(number[1]) { //not decimal (implied) format
			return 0, fmt.Errorf("invalid number")
		}
	}
	parsed, err := strconv.ParseUint(trimmedNumber, base, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", number)
	}
	return uint16(parsed), nil

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
	format, _, err := getOutputFormatAndSize(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	if format == "i" {
		fmt.Println("Cannot print as an instruction")
		return
	}
	var value uint8
	switch strings.ToUpper(args[1]) {
	case "PC":
		fmt.Printf("0x%04X\n", cpu.PC)
		return
	case "X":
		value = cpu.X
	case "Y":
		value = cpu.Y
	case "A":
		value = cpu.A
	case "S":
		value = cpu.S
	case "CF":
		value = boolToUint8(cpu.CF)
	case "ZF":
		value = boolToUint8(cpu.ZF)
	case "IF":
		value = boolToUint8(cpu.IF)
	case "DF":
		value = boolToUint8(cpu.DF)
	case "OF":
		value = boolToUint8(cpu.OF)
	case "NF":
		value = boolToUint8(cpu.NF)
	default:
		fmt.Println("Can't print " + args[1])
		return
	}
	fmt.Printf("%s:\t"+format+"\n", strings.ToUpper(args[1]), value)
}

// uses disassembler to print the current instruction pointed to by the program counter
func printCurrentInstr() {
	instr, _ := nes.DiassembleInstruction(bus, cpu.PC)
	fmt.Printf("0x%04X:\t%s |\tCycles left in Instruction: %d\n", cpu.PC, instr, cpu.Cycles)
}

func main() {

	//if the user didn't provide a binary file as a command line argument
	if len(os.Args) == 1 {
		binaryLoaded := false
		for !binaryLoaded {
			fmt.Println("Enter path to binary:")
			var path string
			fmt.Scanln(&path)
			binaryLoaded = loadBinary(path)
		}
	} else if len(os.Args) == 2 {
		if !loadBinary(os.Args[1]) {
			fmt.Println(os.Args[1] + " could not be loaded")
			os.Exit(1)
		}
	}

	cpu.Reset()
	fmt.Println("Program Loaded.\nAwaiting Input...")
	scanner := bufio.NewScanner(os.Stdin)
	input := ""
	for {
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
			for cpu.Cycles > 0 {
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
		} else {
			fmt.Println("Invalid Command")
		}

	}
}
