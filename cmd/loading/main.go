package main

import (
	"fmt"
	"time"
)

func main() {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	for i := 0; i < 50; i++ {
		// Use \r to return to beginning of line, no \n to stay on same line
		fmt.Printf("\r%s Loading...", spinner[i%len(spinner)])
		time.Sleep(100 * time.Millisecond)
	}

	// Clear the line and print completion
	fmt.Printf("\r● Completed!         \n") // Extra spaces to clear old text
}
