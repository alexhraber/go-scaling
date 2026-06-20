package main

import "fmt"

type ModuleStatus struct {
	Learner   string
	Module    string
	Completed bool
	Score     int
}

func main() {
	modules := []string{"hello go", "variables and types", "functions and control flow"}
	modules = append(modules, "slices maps and structs")
	modules = append(modules, "errors and return values")

	fmt.Println("modules")
	for _, module := range modules {
		fmt.Println(module)
	}

	scores := map[string]int{
		"hello go":                   1,
		"variables and types":        2,
		"functions and control flow": 3,
		"slices maps and structs":    4,
	}

	fmt.Println("function score", scores["functions and control flow"])
	fmt.Println("slices score", scores["slices maps and structs"])

	status := ModuleStatus{
		Learner:   "Ada",
		Module:    "slices maps and structs",
		Completed: false,
		Score:     95,
	}

	fmt.Println(status.Learner)
	fmt.Println(status.Module)
	fmt.Println(status.Completed)
	fmt.Println(status.Score)
}
