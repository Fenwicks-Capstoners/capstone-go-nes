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

func boolToUint8(x bool) uint8 {
	if x {
		return 1
	} else {
		return 0
	}
}

func printMemCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Missing required argument.")
		return
	}
	if len(args) > 2 {
		fmt.Println("Too many arguments provided.")
		return
	}
	var address uint16
	//get address
	if strings.HasPrefix(args[1], "0x") {
		val, error := strconv.ParseUint(args[1][2:], 16, 16)
		if error != nil {
			fmt.Println(args[1] + " is not a valid number. Format: 0x1234")
			return
		}
		address = uint16(val)
	} else if strings.ToUpper(args[1]) == "PC" {
		address = cpu.PC
	} else {
		fmt.Println("Missing required argument")
		return
	}
	//get format
	if len(args[0]) == 1 {
		fmt.Printf("%s:\t%d\n", args[1], cpu.Bus.GetByte(address))
		return
	} else if strings.HasPrefix(args[0], "x/") && len(args[0]) >= 3 {
		//get format specifier
		var format = "%d"
		if args[0][len(args[0])-1] == 'x' || args[0][len(args[0])-1] == 'X' {
			format = "%02X"
		} else if args[0][2] == 'i' {
			fmt.Printf("%s:\t%s\n", args[1], nes.DiassembleInstruction(bus, address))
			return
		} else if args[0][len(args[0])-1] < '0' || args[0][len(args[0])-1] > '9' {
			fmt.Println("Invalid format specifier")
			return
		}
		num := args[0][2:len(args[0])] //number of bytes to print
		if args[0][len(args[0])-1] == 'x' || args[0][len(args[0])-1] == 'X' {
			num = num[:len(num)-1]
		}
		numBytes, err := strconv.ParseInt(num, 10, 32)
		if err != nil {
			fmt.Println("Invalid number of bytes specified")
			return
		}
		value := ""
		for i := 0; i < int(numBytes); i++ {
			value += fmt.Sprintf(format+" ", cpu.Bus.GetByte(address+uint16(i)))
		}
		fmt.Printf("0x%04X-0x%04X:\t%s\n", address, address+uint16(numBytes), value)

	} else {
		fmt.Printf("%s:\t%d\n", args[1], cpu.Bus.GetByte(address))
	}
}

// Prints cpu register by name
// command is p <register name>
func printCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Missing required argument.")
		return
	}
	if len(args) > 2 {
		fmt.Println("Too many arguments provided.")
		return
	}
	format := "%d"
	if strings.HasPrefix(args[0], "p/") && len(args[0]) >= 3 {
		switch strings.ToLower(args[0][2:]) {
		case "x":
			format = "%02X"
		default:
			fmt.Println("Invalid format specifier. Valid Option: 'x'")
			return
		}
	}
	var value uint8
	switch strings.ToUpper(args[1]) {
	case "PC":
		fmt.Printf("%04X\n", cpu.PC)
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
