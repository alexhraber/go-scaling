# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The `-message` Value In The `docker run` Command

Edit the `docker run` command so `-message` uses different text.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-010-exercise .
docker rm -f go-scaling-module-010-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-010-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-010-exercise -message "exercise configuration"
docker logs -f go-scaling-module-010-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

Observe the `/config` JSON response.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-010-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Change The Default Message In `main_exercise.go`

Edit `main_exercise.go` and change the default message passed to `flag.String`.

Rebuild the exercise image.

Then run without passing `-message` so the program uses the default:

```bash
docker rm -f go-scaling-module-010-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-010-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-010-exercise
docker logs -f go-scaling-module-010-exercise &
LOG_PID=$!
curl http://localhost:8080/config
docker stop go-scaling-module-010-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Observe the default message in the `/config` JSON response.

## 3. Add Another Environment Variable Configuration Value

Edit `main_exercise.go` and add another simple configuration value from an environment variable.

Include a default value when the environment variable is not set.

Return the new configuration value from `/config`.

Rebuild and rerun the exercise image, then observe the updated `/config` JSON response.
