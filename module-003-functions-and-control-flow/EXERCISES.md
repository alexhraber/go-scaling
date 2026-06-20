# Exercises

## 1. Add Another Function That Accepts A `string` And Returns A New Message

Define another function that receives a `string`.

Return a new message from that function in `main_exercise.go` and print it with `fmt.Println`.

## 2. Change The `if` Condition And Observe Which Branch Runs

Change the condition inside `scoreMessage`.

From the repo root, build and run the exercise image:

```bash
docker build --target runtime-base -t go-scaling:runtime .
docker build -f module-003-functions-and-control-flow/Dockerfile_exercise -t go-scaling:module-003-exercise .
docker run --rm go-scaling:module-003-exercise
```

Observe whether the `if` branch or the `else` branch runs.

## 3. Change The `for` Loop Count And Observe The Output

Change the loop so it prints a different number of attempts.

Build and run the exercise image again and observe the output.
