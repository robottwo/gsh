package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("Welcome to gsh! Type 'exit' to quit.")
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display prompt
		fmt.Print("gsh> ")

		// Read user input
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Exit condition
		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Execute command
		if input != "" {
			args := strings.Split(input, " ")
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		}
	}
}
