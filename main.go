package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/fatih/color"
)

type Process struct {
	PID  string
	User string
	CPU  string
	Mem  string
	CMD  string
}

type Table struct {
	Headers []string
	Rows    [][]string
	Width   []int
}

// Color scheme
var (
	headerColor    = color.New(color.FgHiCyan, color.Bold, color.BgBlue)
	borderColor    = color.New(color.FgHiBlue, color.Bold)
	highCPUColor   = color.New(color.FgHiRed, color.Bold)
	mediumCPUColor = color.New(color.FgHiYellow, color.Bold)
	lowCPUColor    = color.New(color.FgHiGreen)
	highMemColor   = color.New(color.FgHiMagenta, color.Bold)
	userColor      = color.New(color.FgHiCyan)
	pidColor       = color.New(color.FgHiWhite, color.Bold)
	cmdColor       = color.New(color.FgWhite)
	promptColor    = color.New(color.FgHiGreen, color.Bold)
	errorColor     = color.New(color.FgHiRed, color.Bold, color.BgBlack)
	successColor   = color.New(color.FgHiGreen, color.Bold, color.BgBlack)
	infoColor      = color.New(color.FgHiBlue, color.Bold)
)

func getProcesses() ([]Process, error) {
	cmd := exec.Command("ps", "-eo", "pid,user,%cpu,%mem,comm", "--sort=-%cpu")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var processes []Process

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		p := Process{
			PID:  fields[0],
			User: fields[1],
			CPU:  fields[2],
			Mem:  fields[3],
			CMD:  fields[4],
		}
		processes = append(processes, p)
	}
	return processes, nil
}

func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    make([][]string, 0),
		Width:   make([]int, len(headers)),
	}
}

func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
	// Update column widths
	for i, cell := range row {
		if i < len(t.Width) {
			cleanCell := removeAnsiCodes(cell)
			if len(cleanCell) > t.Width[i] {
				t.Width[i] = len(cleanCell)
			}
		}
	}
	for i, header := range t.Headers {
		if len(header) > t.Width[i] {
			t.Width[i] = len(header)
		}
	}
}

func removeAnsiCodes(s string) string {
	inEscape := false
	result := ""
	for _, char := range s {
		if char == '\033' {
			inEscape = true
			continue
		}
		if inEscape && char == 'm' {
			inEscape = false
			continue
		}
		if !inEscape {
			result += string(char)
		}
	}
	return result
}

func (t *Table) Render() {
	// Print top border
	t.printBorder("┌", "┬", "┐", "─")
	
	// Print header
	fmt.Print("│ ")
	for i, header := range t.Headers {
		fmt.Printf("%s", headerColor.Sprintf("%-*s", t.Width[i], header))
		if i < len(t.Headers)-1 {
			fmt.Print(" │ ")
		}
	}
	fmt.Println(" │")
	
	t.printBorder("├", "┼", "┤", "─")
	
	// Print rows
	for _, row := range t.Rows {
		fmt.Print("│ ")
		for i, cell := range row {
			if i < len(t.Width) {
				cleanCell := removeAnsiCodes(cell)
				padding := t.Width[i] - len(cleanCell)
				fmt.Printf("%s%s", cell, strings.Repeat(" ", padding))
				if i < len(row)-1 {
					fmt.Print(" │ ")
				}
			}
		}
		fmt.Println(" │")
	}
	
	t.printBorder("└", "┴", "┘", "─")
}

func (t *Table) printBorder(left, middle, right, horizontal string) {
	fmt.Print(borderColor.Sprint(left))
	for i, width := range t.Width {
		fmt.Print(borderColor.Sprint(strings.Repeat(horizontal, width+2)))
		if i < len(t.Width)-1 {
			fmt.Print(borderColor.Sprint(middle))
		}
	}
	fmt.Println(borderColor.Sprint(right))
}

func formatCPU(cpu string) string {
	cpuVal, err := strconv.ParseFloat(cpu, 64)
	if err != nil {
		return cpu
	}
	
	if cpuVal > 50.0 {
		return highCPUColor.Sprintf("%.1f%%", cpuVal)
	} else if cpuVal > 20.0 {
		return mediumCPUColor.Sprintf("%.1f%%", cpuVal)
	} else {
		return lowCPUColor.Sprintf("%.1f%%", cpuVal)
	}
}

func formatMemory(mem string) string {
	memVal, err := strconv.ParseFloat(mem, 64)
	if err != nil {
		return mem
	}
	
	if memVal > 10.0 {
		return highMemColor.Sprintf("%.1f%%", memVal)
	} else {
		return cmdColor.Sprintf("%.1f%%", memVal)
	}
}

func formatCommand(cmd string) string {
	if len(cmd) > 25 {
		return cmdColor.Sprint(cmd[:22] + "...")
	}
	return cmdColor.Sprint(cmd)
}

func printProcessTable(processes []Process) {
	fmt.Print("\033[2J\033[H") 
	

	
	fmt.Printf("\n%s %s\n\n", 
		infoColor.Sprint("Total Processes:"), 
		successColor.Sprintf("%d", len(processes)))
	
	// Create custom table
	table := NewTable([]string{"PID", "USER", "CPU%", "MEM%", "COMMAND"})
	
	// Add processes to table (limit to top 15 for better display)
	limit := 15
	if len(processes) < limit {
		limit = len(processes)
	}
	
	for i := 0; i < limit; i++ {
		p := processes[i]
		row := []string{
			pidColor.Sprint(p.PID),
			userColor.Sprint(p.User),
			formatCPU(p.CPU),
			formatMemory(p.Mem),
			formatCommand(p.CMD),
		}
		table.AddRow(row)
	}
	
	table.Render()
	
	// Print legend
	fmt.Printf("\n%s\n", infoColor.Sprint("Legend:"))
	fmt.Printf("  %s High CPU (>50%%)  %s Medium CPU (>20%%)  %s Low CPU\n",
		highCPUColor.Sprint("RED"), mediumCPUColor.Sprint("YELLOW"), lowCPUColor.Sprint("GREEN"))
	fmt.Printf("  %s High Memory (>10%%)  %s Process terminated successfully\n",
		highMemColor.Sprint("PURPLE"), successColor.Sprint("SUCCESS"))
}

func killProcess(pid string) error {
	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}

func printHelp() {
	help := `
╔════════════════════════════════════════════════════════════════════════════╗
║                                COMMANDS                                    ║
╠════════════════════════════════════════════════════════════════════════════╣
║  list                    │  Show running processes                         ║
║  kill <pid>              │  Terminate process by PID                       ║
║  refresh                 │  Auto-refresh process list                      ║
║  top [n]                 │  Show top N processes (default: 15)             ║
║  help                    │  Show this help menu                            ║
║  clear                   │  Clear the screen                               ║
║  exit                    │  Exit htop-go                                   ║
╚════════════════════════════════════════════════════════════════════════════╝`
	
	fmt.Println(infoColor.Sprint(help))
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func refreshMode() {
	fmt.Println(infoColor.Sprint("Auto-refresh mode (Press Ctrl+C to stop)"))
	
	for {
		clearScreen()
		procs, err := getProcesses()
		if err != nil {
			errorColor.Printf("Error fetching processes: %v\n", err)
			return
		}
		printProcessTable(procs)
		fmt.Printf("\n%s %s\n", 
			infoColor.Sprint("Last updated:"), 
			time.Now().Format("15:04:05"))
		fmt.Println(infoColor.Sprint("Press Ctrl+C to stop auto-refresh"))
		time.Sleep(2 * time.Second)
	}
}

func showTopProcesses(n int) {
	procs, err := getProcesses()
	if err != nil {
		errorColor.Printf("Error fetching processes: %v\n", err)
		return
	}
	
	if n > len(procs) {
		n = len(procs)
	}
	
	clearScreen()
	fmt.Printf("%s %s\n\n", 
		infoColor.Sprint("Top"), 
		successColor.Sprintf("%d", n))
	
	table := NewTable([]string{"RANK", "PID", "USER", "CPU%", "MEM%", "COMMAND"})
	
	for i := 0; i < n; i++ {
		p := procs[i]
		rank := fmt.Sprintf("#%d", i+1)
		if i < 3 {
			rank = fmt.Sprintf("[%d]", i+1)
		}
		
		row := []string{
			rank,
			pidColor.Sprint(p.PID),
			userColor.Sprint(p.User),
			formatCPU(p.CPU),
			formatMemory(p.Mem),
			formatCommand(p.CMD),
		}
		table.AddRow(row)
	}
	
	table.Render()
}

func commandLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Printf("\n%s", promptColor.Sprint("htop-go ⚡ > "))
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		args := strings.Fields(input)
		
		if len(args) == 0 {
			continue
		}
		
		switch args[0] {
		case "list", "ls":
			procs, err := getProcesses()
			if err != nil {
				errorColor.Printf("Error fetching processes: %v\n", err)
				continue
			}
			printProcessTable(procs)
			
		case "kill":
			if len(args) < 2 {
				errorColor.Println("ERROR: Usage: kill <pid>")
				continue
			}
			if _, err := strconv.Atoi(args[1]); err != nil {
				errorColor.Println("ERROR: Invalid PID - must be a number")
				continue
			}
			
			fmt.Printf("%s %s", infoColor.Sprint("WARNING: Are you sure you want to kill process"), args[1])
			fmt.Print("? (y/N): ")
			scanner.Scan()
			confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
			
			if confirm == "y" || confirm == "yes" {
				err := killProcess(args[1])
				if err != nil {
					errorColor.Printf("ERROR: Failed to kill process %s: %v\n", args[1], err)
				} else {
					successColor.Printf("SUCCESS: Successfully killed process %s\n", args[1])
				}
			} else {
				fmt.Println(infoColor.Sprint("CANCELLED: Kill operation cancelled"))
			}
			
		case "refresh", "r":
			refreshMode()
			
		case "top":
			n := 15 // default
			if len(args) > 1 {
				if num, err := strconv.Atoi(args[1]); err == nil && num > 0 {
					n = num
				}
			}
			showTopProcesses(n)
			
		case "help", "h", "?":
			printHelp()
			
		case "clear", "cls":
			clearScreen()
			
		case "exit", "quit", "q":
			fmt.Println(successColor.Sprint("Thanks for using htop-go! Goodbye!"))
			return
			
		default:
			errorColor.Printf("ERROR: Unknown command: %s\n", args[0])
			fmt.Println(infoColor.Sprint("INFO: Type 'help' to see available commands"))
		}
	}
}

func printWelcomeBanner() {
	clearScreen()
	
	banner := `

  ██╗  ██╗████████╗ ██████╗ ██████╗        ██████╗  ██████╗  
  ██║  ██║╚══██╔══╝██╔═══██╗██╔══██╗      ██╔════╝ ██╔═══██╗ 
  ███████║   ██║   ██║   ██║██████╔╝█████╗██║  ███╗██║   ██║ 
  ██╔══██║   ██║   ██║   ██║██╔═══╝ ╚════╝██║   ██║██║   ██║ 
  ██║  ██║   ██║   ╚██████╔╝██║           ╚██████╔╝╚██████╔╝ 
  ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝            ╚═════╝  ╚═════╝  
`
	
	fmt.Println(headerColor.Sprint(banner))
	fmt.Printf("\n%s\n", infoColor.Sprint("Ready to monitor your system! Type 'help' for commands or 'list' to start."))
}

func main() {
	printWelcomeBanner()
	commandLoop()
}