# Module 005: Errors And Return Values

This module introduces the next primitive of a Go program: programs report failure as values, and callers decide what to do next.

## A Return Value

A return value is a value a function gives back to the code that called it.

```go
return module + " is ready", nil
```

This function returns a message and an error value.

## Multiple Return Values

Go functions can return more than one value.

```go
func moduleMessage(module string, score int) (string, error)
```

This function returns a `string` result and an `error`.

## `error`

An `error` is a value that describes a failure.

```go
return "", errors.New("module name is required")
```

The caller receives the error and decides what to do next.

## `nil`

`nil` means there is no error value.

```go
return module + " is ready", nil
```

This is the no-error path.

## `if err != nil`

`if err != nil` checks whether an error happened.

```go
if err != nil {
	fmt.Println("failure:", err)
	return
}
```

If `err` is not `nil`, the function handles the failure.

## Early Return

An early return stops the current function before the success path runs.

This keeps failure handling simple: handle the error, return, and let the rest of the function stay focused on success.

## Success Path

The success path runs when `err` is `nil`.

```go
fmt.Println("success:", message)
```

The caller prints the returned message.

## Failure Path

The failure path runs when `err` contains an error value.

```go
fmt.Println("failure:", err)
```

The caller prints the error and returns early.

## Build The Shared Runtime Base

From the repo root, run:

```bash
docker build --target runtime-base -t go-scaling:runtime .
```

This builds the small runtime image that module Dockerfiles use after compiling a Go binary.

## Build The Module Image

From the repo root, run:

```bash
docker build -f module-005-errors-and-return-values/Dockerfile -t go-scaling:module-005 .
```

The Dockerfile compiles `module-005-errors-and-return-values/main.go` into a binary and copies that binary into the runtime image.

## Run The Module Image

From the repo root, run:

```bash
docker run --rm go-scaling:module-005
```

The container starts the binary as an operating-system process and prints the program output.

## What Changed From Module 004

Module 004 introduced grouped data with slices, maps, and structs.

This module uses functions to return either a normal value or an error value. The caller checks the error and chooses the success path or the failure path.

## First-Principles Chain

```text
source code → typed values → grouped data → functions → return values → errors → control flow → compiler checks → binary → operating-system process
```

Source code contains typed values. Grouped data organizes those values. Functions return values. Errors report failure as values. Control flow chooses the next path. The compiler checks the program before building a binary. The operating system starts that binary as a process.

## Why Errors And Return Values Matter For Scaling Later

Larger programs need explicit failure paths.

Return values let functions pass useful results back to callers. Error values let functions report what went wrong without hiding failure.
