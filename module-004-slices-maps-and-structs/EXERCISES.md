# Exercises

## 1. Add Another Value To The Slice And Print It

Add another module name to `modules` in `main_exercise.go`.

The shared runtime image is built once from the repo root earlier in the learning flow.

From the repo root, enter this module directory, then build and run the exercise image:

```bash
cd module-004-slices-maps-and-structs
docker build -f Dockerfile_exercise -t go-scaling:module-004-exercise .
docker run --rm go-scaling:module-004-exercise
```

Observe the printed list.

## 2. Add Another Key/Value Pair To The Map And Read It

Add another module score to `scores`.

Read that value with its key and print it with `fmt.Println`.

## 3. Add Another Field To The Struct And Print It

Add another field to `ModuleStatus`.

Set that field when creating `status`, then print it with `fmt.Println`.
