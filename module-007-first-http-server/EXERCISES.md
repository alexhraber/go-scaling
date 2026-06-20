# Exercises

## 1. Change The Response Text

Edit `main_exercise.go` so the handler writes different response text.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-007-exercise .
docker run --rm -p 8080:8080 go-scaling:module-007-exercise
```

In a second terminal, call the server:

```bash
curl http://localhost:8080/
```

Observe the output.

## 2. Change The Startup Message

Edit `main_exercise.go` so the startup message printed to standard output is different.

Rebuild and rerun the exercise image, then observe the new message when the server starts.

## 3. Add Another Response Line

Edit `main_exercise.go` and add a second plain text line to the HTTP response without adding another route.

Rebuild and rerun the exercise image, then observe the updated `curl` output.
