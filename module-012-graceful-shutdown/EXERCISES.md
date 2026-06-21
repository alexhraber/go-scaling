# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Shutdown Timeout

Edit `main_exercise.go` and change the shutdown timeout passed to `context.WithTimeout`.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-012-exercise .
docker rm -f go-scaling-module-012-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-012-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-012-exercise -message "exercise graceful shutdown"
docker logs -f go-scaling-module-012-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

Observe the request logs.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-012-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Observe the shutdown log output.

## 2. Change The Shutdown Log Messages

Edit `main_exercise.go` and change the startup, shutdown start, and shutdown complete log messages so they are easy to distinguish.

Rebuild and rerun the exercise image, then observe the updated log output.

## 3. Add A `/version` Route

Edit `main_exercise.go` and add a simple `/version` route that returns JSON.

Make the route log its method, path, status, and duration before graceful shutdown.

After adding `/version`, call it from the same terminal:

```bash
curl http://localhost:8080/version
```

Rebuild and rerun the exercise image, then observe the new JSON response, request log line, and shutdown log output.
