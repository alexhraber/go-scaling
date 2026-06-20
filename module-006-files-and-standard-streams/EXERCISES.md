# Exercises

## 1. Change The Standard Input Text

Edit `main_exercise.go` so the program handles different stdin text. (hint: does the script need to change?)

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-006-exercise .
printf "exercise input\n" | docker run --rm -i go-scaling:module-006-exercise
```

Observe how the stdout file output changes.

## 2. Change The Temporary File Path

Edit `main_exercise.go` and change the temporary file path.

Rebuild and rerun the exercise image to confirm the program still writes the file and reads it back.

Observe the output.

## 3. Add More Stream Output

Edit `main_exercise.go` and add another stdout message and another stderr message.

Rebuild and rerun the exercise image to observe both streams when the container runs.
