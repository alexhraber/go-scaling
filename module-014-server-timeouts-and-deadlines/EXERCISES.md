# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Handler Deadline So `/work` Completes

Edit `main_exercise.go` and change the deadline passed to `context.WithTimeout` so it is longer than the simulated work duration.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-014-exercise .
docker rm -f go-scaling-module-014-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-014-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-014-exercise
docker logs -f go-scaling-module-014-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl -i http://localhost:8080/work
curl http://localhost:8080/metrics
```

Observe `work_completed_total` in the `/metrics` JSON response.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-014-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Change The Simulated Work Duration So `/work` Times Out Again

Edit `main_exercise.go` and change the simulated work duration so it is longer than the handler deadline.

Rebuild and rerun the exercise image, then observe `work_timed_out_total` in the `/metrics` JSON response.

## 3. Add A `/healthz` Counter

Edit `main_exercise.go` and add a new primitive counter for `/healthz` requests.

Expose that counter from `/metrics`.

Rebuild and rerun the exercise image, then observe the updated `/metrics` JSON response.
