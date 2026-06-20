# Module 002: Variables, Values, And Types

This module introduces the next primitive of a Go program: values have types, and types let programs make promises before they run.

## A Value

A value is a piece of data a program can use.

`"Hello from Go."` is a value. `42` is a value. `true` is a value.

## A Type

A type names the kind of value a program is allowed to use.

Go checks types before the program runs. That helps the compiler catch mistakes early.

## `string`

A `string` is text.

Examples include `"Go"`, `"learner"`, and `"values have types"`.

## `int`

An `int` is a whole number.

Examples include `1`, `2`, and `42`.

## `bool`

A `bool` is either `true` or `false`.

Use a `bool` for a yes-or-no value.

## `var`

`var` declares a variable with a name and a type.

```go
var learnerName string
```

This creates a `string` variable named `learnerName`.

## `:=`

`:=` declares a variable and assigns a value at the same time.

```go
topic := "values have types"
```

Go uses the value on the right side to infer the variable's type.

## Zero Values

When a variable is declared with `var` but no value, Go gives it a zero value.

The zero value for `string` is empty text. The zero value for `int` is `0`. The zero value for `bool` is `false`.

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-002 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker run --rm go-scaling:module-002
```

The container starts the binary as an operating-system process and prints the program output.

## What Changed From Module 001

Module 001 printed one string value.

This module prints several values: a `string`, an `int`, and a `bool`. It also shows two ways to declare variables: `var` and `:=`.

## First-Principles Chain

```text
source code → typed values → compiler checks → binary → operating-system process
```

Source code contains typed values. The compiler checks those types before building a binary. The operating system starts that binary as a process.

## Why Types Matter For Scaling Later

Larger programs pass values between more files, functions, packages, and services.

Types make those values explicit. They let the compiler reject many mistakes before a larger system is running.
