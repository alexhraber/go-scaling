# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Fixed-Window Limit

Edit `main_exercise.go` and change the fixed-window limit so more `/work` requests are admitted before rejection.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-015-exercise .
docker rm -f go-scaling-module-015-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-015-exercise -p 8080:8080 -e APP_PORT=8080 go-scaling:module-015-exercise
docker logs -f go-scaling-module-015-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl -i http://localhost:8080/work
curl -i http://localhost:8080/work
curl -i http://localhost:8080/work
curl http://localhost:8080/metrics
```

Observe how many `/work` requests are admitted before one is rejected.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-015-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Change The Fixed-Window Duration

Edit `main_exercise.go` and change the fixed-window duration.

After changing the duration and rebuilding, wait for the next window:

```bash
sleep 11
curl -i http://localhost:8080/work
curl http://localhost:8080/metrics
```

Observe how waiting for the next window changes the result.

## 3. Add A `/healthz` Counter

Edit `main_exercise.go` and add a new primitive counter for `/healthz` requests.

Expose that counter from `/metrics`.

Rebuild and rerun the exercise image, then observe the updated `/metrics` JSON response.
