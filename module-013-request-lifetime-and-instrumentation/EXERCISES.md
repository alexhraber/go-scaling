# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Simulated Work Duration

Edit `main_exercise.go` and change the duration passed to `time.NewTimer` in the `/slow` handler.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-013-exercise .
docker rm -f go-scaling-module-013-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-013-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-013-exercise
docker logs -f go-scaling-module-013-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl http://localhost:8080/slow
curl --max-time 1 http://localhost:8080/slow || true
curl http://localhost:8080/metrics
```

Observe how normal completion and cancellation behave differently.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-013-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Add A `/healthz` Counter

Edit `main_exercise.go` and add a new primitive counter for `/healthz` requests.

Expose that counter from `/metrics`.

Rebuild and rerun the exercise image, then observe the updated `/metrics` JSON response.

## 3. Add A `/readyz` Route

Edit `main_exercise.go` and add a small `/readyz` route that returns JSON.

Increment its own counter and include that counter in `/metrics`.

After adding `/readyz`, call it from the same terminal:

```bash
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

Rebuild and rerun the exercise image, then observe the new JSON response, request log line, and metrics output.
