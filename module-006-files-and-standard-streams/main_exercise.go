package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "stdout: module 006 is running")
	fmt.Fprintln(os.Stderr, "stderr: this is an error-style message")

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read stdin:", err)
		return
	}

	notePath := "/tmp/module-006-note.txt"
	if err := os.WriteFile(notePath, input, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write file:", err)
		return
	}

	note, err := os.ReadFile(notePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read file:", err)
		return
	}

	fmt.Println("stdin input:")
	fmt.Print(string(input))
	fmt.Println("file path:", notePath)
	fmt.Println("file contents:")
	fmt.Print(string(note))
}
