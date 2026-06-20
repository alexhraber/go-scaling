# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Configured `-message` Value

Edit the `docker run` command so `-message` uses different text.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-011-exercise .
docker rm -f go-scaling-module-011-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-011-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-011-exercise -message "exercise logging"
docker logs -f go-scaling-module-011-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

Observe both the `/config` JSON response and the request log line.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-011-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Add The Configured Message To The Log Line

Edit `main_exercise.go` and change the log line format so it includes the configured message.

Rebuild and rerun the exercise image, then observe the updated request log output.

## 3. Add A `/version` Route

Edit `main_exercise.go` and add a simple `/version` route that returns JSON.

Make the route log its method, path, status, and duration.

After adding `/version`, call it from the same terminal:

```bash
curl http://localhost:8080/version
```

Rebuild and rerun the exercise image, then observe the new JSON response and request log line.
