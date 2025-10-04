package selectx

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectWithPrompt provides a simple number-based selection UI
func SelectWithPrompt(items []string, prompt string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select from")
	}

	// Auto-select if only one item
	if len(items) == 1 {
		return 0, nil
	}

	// Display items with numbers
	fmt.Fprintf(os.Stderr, "%s:\n", prompt)
	for i, item := range items {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, item)
	}
	fmt.Fprintf(os.Stderr, "\nSelect number (1-%d, or q to quit): ", len(items))

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Check for cancellation
	if input == "q" || input == "Q" || input == "" {
		return -1, fmt.Errorf("selection cancelled")
	}

	// Convert to number
	num, err := strconv.Atoi(input)
	if err != nil {
		return -1, fmt.Errorf("invalid input: %s", input)
	}

	// Validate range
	if num < 1 || num > len(items) {
		return -1, fmt.Errorf("number out of range: %d (expected 1-%d)", num, len(items))
	}

	return num - 1, nil
}
