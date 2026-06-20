package main

import "fmt"

func scoreMessage(name string, score int) string {
	if score >= 70 {
		return name + " passed"
	} else {
		return name + " needs more practice"
	}
}

func gameResult(participant string, score int) string {
	if score >= 80 {
		return participant + " won by default"
	} else {
		return participant + " lost because they didn't have enough points"
	}
}

func main() {
	learner := "Ada"
	score := 82
	team := "Lakers"

	result := scoreMessage(learner, score)
	fmt.Println(result)

	result = gameResult(team, score)
	fmt.Println(result)

	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Println("attempt", attempt)
	}
}
