package main

import "fmt"

func scoreMessage(name string, score int) string {
	if score >= 70 {
		return name + " passed"
	} else {
		return name + " needs more practice"
	}
}

func main() {
	learner := "Ada"
	score := 82

	result := scoreMessage(learner, score)
	fmt.Println(result)

	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Println("attempt", attempt)
	}
}
