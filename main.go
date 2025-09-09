package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"postboy/cmd"

	tea "github.com/charmbracelet/bubbletea"
)

func runTUI(filePath string) {
	p := tea.NewProgram(cmd.NewModel(filePath),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runTUIWithFile(filePath string) {
	data, err := cmd.LoadHTTPFile(filePath)
	if err != nil || len(data.Requests) == 0 {
		// Remove postboy.http if exists to force dummy request
		runTUI(filePath)
		return
	}
	runTUI(filePath)
}

func main() {
	if len(os.Args) == 1 {
		runTUI("")
		os.Exit(0)
	}

	if len(os.Args) == 3 && os.Args[1] == "read" {
		runTUIWithFile(os.Args[2])
		os.Exit(0)
	}

	if len(os.Args) == 2 && os.Args[1] == "run" {
		fmt.Println("Error: Missing method and endpoint.")
		fmt.Println("Usage: myclient run <method> <url> [-s] [json_payload]")
		os.Exit(1)
	}

	if len(os.Args) < 4 || strings.ToLower(os.Args[1]) != "run" {
		fmt.Println("Usage: myclient run <method> <url> [-s] [json_payload]")
		os.Exit(1)
	}

	method := strings.ToUpper(os.Args[2])
	url := os.Args[3]

	// Check for -s flag
	simpleOutput := false
	payloadIdx := 4
	if len(os.Args) > 4 && os.Args[4] == "-s" {
		simpleOutput = true
		payloadIdx = 5
	}
	cmd.SendByCLI(method, url, simpleOutput, payloadIdx)
}
