# Module 003: Functions And Control Flow

This module introduces the next primitive of a Go program: programs make decisions and reuse named blocks of behavior.

## A Function

A function is a named block of behavior.

It groups code so the program can reuse it.

## A Function Call

A function call runs a function.

```go
result := scoreMessage(learner, score)
```

This calls `scoreMessage` and stores the returned value in `result`.

## Parameters

Parameters are named values a function receives.

```go
func scoreMessage(name string, score int) string
```

This function receives a `string` named `name` and an `int` named `score`.

## Return Values

A return value is the value a function gives back to the code that called it.

```go
return name + " passed"
```

This returns a `string`.

## `if`

`if` runs code only when a condition is true.

```go
if score >= 70 {
	return name + " passed"
}
```

## `else`

`else` runs when the `if` condition is false.

```go
else {
	return name + " needs more practice"
}
```

## `for`

A `for` loop repeats code.

```go
for attempt := 1; attempt <= 3; attempt++ {
	fmt.Println("attempt", attempt)
}
```

This loop prints three attempts.

## Control Flow

Control flow is the order in which code runs.

Functions choose reusable behavior. `if` and `else` choose between branches. `for` repeats a block of code.

## Run The Source File

From this directory, run:

```bash
go run main.go
```

`go run` compiles the source code and immediately runs the resulting program.

## Build The Binary

From this directory, run:

```bash
go build
```

`go build` compiles the package and writes an executable binary named `module-003-functions-and-control-flow`.

## Run The Binary

After building, run:

```bash
./module-003-functions-and-control-flow
```

## What Changed From Module 002

Module 002 introduced typed values and variables.

This module uses typed values as inputs to a function. The function makes a decision with `if` and `else`, returns a result, and the program repeats output with a `for` loop.

## First-Principles Chain

```text
source code → typed values → functions → control flow → compiler checks → binary → operating-system process
```

Source code contains typed values. Functions organize behavior around those values. Control flow chooses which code runs. The compiler checks the program before building a binary. The operating system starts that binary as a process.

## Why Functions And Control Flow Matter For Scaling Later

Larger programs need behavior that can be named, reused, and checked.

Functions keep behavior in clear places. Control flow lets a program respond to different values without duplicating every step.
