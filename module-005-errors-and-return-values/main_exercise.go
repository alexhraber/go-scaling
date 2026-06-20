package main

import (
	"errors"
	"fmt"
)

func moduleMessage(module string, score int) (string, error) {
	if module == "" {
		return "", errors.New("module name is required")
	}

	if score < 0 {
		return "", errors.New("score cannot be negative")
	}

	return module + " is ready", nil
}

func printModuleStatus(module string, score int) {
	message, err := moduleMessage(module, score)
	if err != nil {
		fmt.Println("failure:", err)
		return
	}

	fmt.Println("success:", message)
}

func main() {
	printModuleStatus("errors and return values", 90)
	printModuleStatus("", 90)
}
