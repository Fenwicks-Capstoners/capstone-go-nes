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

func boolToUint16(x bool) uint16 {
	if x {
		return 1
	} else {
		return 0
	}
}

// cli print command
func printCmd(args []string) {
	if len(args) == 1 {
		fmt.Println("Missing required argument")
		return
	}
	var target string
	var value uint16
	if len(args) == 2 {
		target = args[1]
	} else {
		target = args[2]
	}

	switch strings.ToLower(target) {
	case "pc":
		value = cpu.PC
	case "x":
		value = uint16(cpu.X)
	case "y":
		value = uint16(cpu.Y)
	case "a":
		value = uint16(cpu.A)
	case "s":
		value = uint16(cpu.S)
	case "cf":
		value = boolToUint16(cpu.CF)
	case "zf":
		value = boolToUint16(cpu.ZF)
	case "if":
		value = boolToUint16(cpu.IF)
	case "df":
		value = boolToUint16(cpu.DF)
	case "of":
		value = boolToUint16(cpu.OF)
	case "nf":
		value = boolToUint16(cpu.NF)
	case "operand":
		value = cpu.Operand
	default:
		if strings.HasPrefix(strings.ToLower(target), "0x") {
			v, error := strconv.ParseUint(target[2:], 16, 16)
			if error != nil {
				fmt.Println(target + " is not a valid number " + error.Error())
				return
			}
			value = uint16(v)

		} else {
			fmt.Println("Cannot print", target)
			return
		}
	}

	format := "%X"
	if len(args) > 2 {
		format = args[1]
	}
	if format == "-i" {
		fmt.Println(nes.DiassembleInstruction(bus, value))
	} else if format == "-16" {
		fmt.Printf("%s:\t%04X\n", target, cpu.Get2Bytes(value))

	} else {
		fmt.Printf(target+":\t"+format+"\n", cpu.Bus.GetByte(value))
	}
}

// uses disassembler to print the current instruction pointed to by the program counter
func printCurrentInstr() {
	fmt.Printf("%s |\tCycles left in Instruction: %d\n", nes.DiassembleInstruction(bus, cpu.PC), cpu.Cycles)
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
	for input != "quit" {
		scanner.Scan()
		input = scanner.Text()
		tokens := strings.Fields(input)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "clear":
			fmt.Print("\033[H\033[2J")
		case "print":
			printCmd(tokens)
		case "nc":
			cpu.Clock()
			printCurrentInstr()
		case "ni":
			cpu.Clock()
			for cpu.Cycles > 0 {
				cpu.Clock()
			}
			printCurrentInstr()
		case "cur":
			printCurrentInstr()
		default:
			fmt.Println("Invalid Command")
		}

	}
}
