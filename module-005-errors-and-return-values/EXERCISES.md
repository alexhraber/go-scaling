# Exercises

## 1. Change The Invalid Input Case In `main_exercise.go` And Observe The Error Output

Change the invalid input passed to `printModuleStatus`.

From the repo root, build and run the exercise image:

```bash
docker build --target runtime-base -t go-scaling:runtime .
docker build -f module-005-errors-and-return-values/Dockerfile_exercise -t go-scaling:module-005-exercise .
docker run --rm go-scaling:module-005-exercise
```

Observe the printed error output.

## 2. Add Another Validation Rule That Returns An Error

Add another `if` check inside `moduleMessage`.

Return a new error when that rule fails.

## 3. Add Another Successful Input And Print The Returned Message

Add another valid call to `printModuleStatus`.

Build and run the exercise image again to observe the success output.
