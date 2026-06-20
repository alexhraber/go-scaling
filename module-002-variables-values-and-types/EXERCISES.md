# Exercises

## 1. Add Another `string` Variable And Print It

Declare another `string` variable in `main_exercise.go`.

Print it with `fmt.Println`.

## 2. Change An `int` Value And Observe The Output

Change the value assigned to `count`.

The shared runtime image is built once from the repo root earlier in the learning flow.

From the repo root, enter this module directory, then build and run the exercise image:

```bash
cd module-002-variables-values-and-types
docker build -f Dockerfile_exercise -t go-scaling:module-002-exercise .
docker run --rm go-scaling:module-002-exercise
```

Observe the output.

## 3. Declare A `bool` Zero Value With `var` And Print It

Declare a `bool` variable with `var` and no assigned value.

Print it with `fmt.Println`.
