# Module 004: Slices, Maps, And Structs

This module introduces the next primitive of a Go program: programs group values so they can model lists, lookups, and records.

## A Slice

A slice is an ordered list of values.

```go
modules := []string{"hello go", "variables and types"}
```

This slice holds `string` values in order.

## Ordered Values

A slice is useful when order matters.

The first module can come before the second module, and a loop can read each value in that order.

## `append`

`append` adds a value to a slice and returns the updated slice.

```go
modules = append(modules, "slices maps and structs")
```

The result is assigned back to `modules`.

## A Map

A map is a keyed lookup.

```go
scores := map[string]int{
	"hello go": 1,
}
```

This map uses a `string` key to find an `int` value.

## Keyed Lookup

A map is useful when a program needs one value by name.

```go
scores["hello go"]
```

This reads the value stored for the key `"hello go"`.

## A Struct

A struct is a named record.

```go
type ModuleStatus struct {
	Learner   string
	Module    string
	Completed bool
}
```

This struct groups related fields under one type name.

## Related Fields

A struct is useful when values belong together.

A learner name, module name, and completion value describe one module status.

## How They Differ

A slice groups values in order.

A map groups values by key.

A struct groups named fields into one record.

## Build The Shared Runtime Base

From the repo root, run:

```bash
docker build --target runtime-base -t go-scaling:runtime .
```

This builds the small runtime image that module Dockerfiles use after compiling a Go binary.

## Build The Module Image

From the repo root, run:

```bash
docker build -f module-004-slices-maps-and-structs/Dockerfile -t go-scaling:module-004 .
```

The Dockerfile compiles `module-004-slices-maps-and-structs/main.go` into a binary and copies that binary into the runtime image.

## Run The Module Image

From the repo root, run:

```bash
docker run --rm go-scaling:module-004
```

The container starts the binary as an operating-system process and prints the program output.

## What Changed From Module 003

Module 003 introduced functions and control flow.

This module keeps control flow with a `for` loop and adds grouped data. A slice stores ordered module names. A map stores scores by module name. A struct stores one learner/module status as a record.

## First-Principles Chain

```text
source code → typed values → grouped data → functions → control flow → compiler checks → binary → operating-system process
```

Source code contains typed values. Grouped data organizes those values. Functions and control flow use those groups. The compiler checks the program before building a binary. The operating system starts that binary as a process.

## Why Grouped Data Matters For Scaling Later

Larger programs work with many related values.

Slices help programs handle ordered collections. Maps help programs find values by key. Structs help programs pass one clear record instead of many separate variables.
