# Exercises

Edit `main_exercise.go` for code changes in these exercises.

## 1. Change The Expected Bearer Token

Edit the `AUTH_TOKEN` value in the `docker run` command.

The old token should fail, and the new token should succeed.

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build and run the exercise image:

```bash
mkdir -p data
docker build -f Dockerfile_exercise -t go-scaling:module-016-exercise .
docker rm -f go-scaling-module-016-exercise >/dev/null 2>&1 || true
docker run --rm -d \
  --name go-scaling-module-016-exercise \
  -p 8080:8080 \
  -e APP_PORT=8080 \
  -e AUTH_TOKEN=module-016-token \
  -e ENCRYPTION_KEY=0123456789abcdef0123456789abcdef \
  -e DATA_FILE=/data/records.enc \
  -v "$PWD/data:/data" \
  go-scaling:module-016-exercise
docker logs -f go-scaling-module-016-exercise &
LOG_PID=$!
```

From the same terminal, call the server:

```bash
curl -i http://localhost:8080/records
curl -i -X POST http://localhost:8080/records \
  -H 'Authorization: Bearer module-016-token' \
  -H 'Content-Type: application/json' \
  -d '{"id":"exercise","text":"encrypted exercise record"}'
curl -i http://localhost:8080/records/exercise \
  -H 'Authorization: Bearer module-016-token'
curl http://localhost:8080/metrics
```

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-016-exercise
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## 2. Add A Missing-Record Counter

Edit `main_exercise.go` and add a new primitive counter for missing records.

Expose that counter from `/metrics`.

Rebuild and rerun the exercise image, then request a record ID that does not exist.

## 3. Add `DELETE /records/{id}`

Edit `main_exercise.go` and add `DELETE /records/{id}`.

The authorized delete should decrypt the file, remove a record, re-encrypt the file, and persist the deletion across a second container instantiation.

After adding `DELETE /records/{id}`, call it from the same terminal:

```bash
curl -i -X DELETE http://localhost:8080/records/exercise \
  -H 'Authorization: Bearer module-016-token'

curl -i http://localhost:8080/records/exercise \
  -H 'Authorization: Bearer module-016-token'

curl http://localhost:8080/metrics
```

Rebuild and rerun the exercise image, then observe the delete response, missing-record response, and updated metrics.
