# Exercises

## 1. Change The Successful `/echo` JSON Response Field Name

Edit `main_exercise.go` so the successful `/echo` response uses a different JSON field name than `received`.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
docker build -f Dockerfile_exercise -t go-scaling:module-009-exercise .
docker rm -f go-scaling-module-009-exercise >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-009-exercise -p 8080:8080 go-scaling:module-009-exercise
docker logs -f go-scaling-module-009-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz

curl -X POST http://localhost:8080/echo \
  -H 'Content-Type: application/json' \
  -d '{"message":"exercise json"}'

curl -i -X POST http://localhost:8080/echo \
  -H 'Content-Type: application/json' \
  -d '{bad json}'

curl -i http://localhost:8080/echo
```

Observe the output.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-009-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Add Another Field To The Request And Response Structs

Edit `main_exercise.go` and add another field to the request struct and the response struct.

Include that field in the returned JSON.

Rebuild and rerun the exercise image, then observe the updated `curl` output.

## 3. Change The Invalid JSON Error Message

Edit `main_exercise.go` and change the invalid JSON error message while keeping status `400`.

Rebuild and rerun the exercise image, then observe the updated `curl -i` output.
