# Exercises

## 1. Change The Printed Message

Edit `main_exercise.go` so the program prints a different message.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-001-exercise .
docker run --rm go-scaling:module-001-exercise
```

## 2. Print Two Lines

Add a second `fmt.Println(...)` call inside `main_exercise.go`.

Build and run the exercise image again and confirm that the program prints two lines.

## 3. Change The Exercise Image Output

Change the text again in `main_exercise.go`.

Build and run the exercise image again to observe the new output.
