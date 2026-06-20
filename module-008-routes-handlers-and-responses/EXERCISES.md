# Exercises

## 1. Change The `/module` Response

Edit `main_exercise.go` so the `/module` handler writes different response text.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-008-exercise .
docker rm -f go-scaling-module-008-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-008-exercise -p 8080:8080 go-scaling:module-008-exercise
docker logs -f go-scaling-module-008-exercise &
LOG_PID=$!
```

From the same terminal, call each route:

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
curl http://localhost:8080/module
curl -i http://localhost:8080/missing
```

Observe the output.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-008-exercise
wait "$LOG_PID" 2>/dev/null || true
```

## 2. Change The `/healthz` Response Body

Edit `main_exercise.go` so the `/healthz` route returns different plain text while keeping status `200`.

Rebuild and rerun the exercise image, then observe the updated route output.

## 3. Change The Not-Found Response Text

Edit `main_exercise.go` so the not-found handler writes different text while keeping the not-found status code.

Rebuild and rerun the exercise image, then observe the updated `curl -i` output.
