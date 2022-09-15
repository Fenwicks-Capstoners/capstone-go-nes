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
	if strings.ToUpper(target) == "PC" {
		value = cpu.PC
	} else if strings.HasPrefix(strings.ToLower(target), "0x") {
		v, error := strconv.ParseUint(target[2:], 16, 16)
		if error != nil {
			fmt.Println(target + " is not a valid number " + error.Error())
			return
		}
		value = uint16(cpu.Bus.GetByte(uint16(v)))
	} else {
		fmt.Println("Cannot print", target)
		return
	}

	format := "%X"
	if len(args) > 2 {
		format = args[1]
	}
	if format == "-i" {
		fmt.Println(nes.DiassembleInstruction(bus, value))
	} else {
		fmt.Printf(target+":\t"+format+"\n", value)
	}
}

func printCurrentInstr() {
	fmt.Printf(nes.DiassembleInstruction(bus, cpu.PC))
}

func main() {

	binaryLoaded := false
	for !binaryLoaded {
		fmt.Println("Enter path to binary:")
		var path string
		fmt.Scanln(&path)
		binaryLoaded = loadBinary(path)
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
		case "print":
			printCmd(tokens)
		case "c":
			cpu.Clock()
		default:
			fmt.Println("Invalid Command")
		}

	}
}
