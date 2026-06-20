# Exercises

## 1. Add Another Value To The Slice And Print It

Add another module name to `modules` in `main_exercise.go`.

From the repo root, build and run the exercise image:

```bash
docker build --target runtime-base -t go-scaling:runtime .
docker build -f module-004-slices-maps-and-structs/Dockerfile_exercise -t go-scaling:module-004-exercise .
docker run --rm go-scaling:module-004-exercise
```

Observe the printed list.

## 2. Add Another Key/Value Pair To The Map And Read It

Add another module score to `scores`.

Read that value with its key and print it with `fmt.Println`.

## 3. Add Another Field To The Struct And Print It

Add another field to `ModuleStatus`.

Set that field when creating `status`, then print it with `fmt.Println`.
