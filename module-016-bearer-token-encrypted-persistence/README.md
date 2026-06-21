# Module 016: Bearer Token Encrypted Persistence

This module introduces the next scaling primitive of a Go program: a service can persist sensitive state across restarts, but every read and write must cross an authorization and encryption boundary.

The core primitive is:

```text
A scalable service can persist sensitive state across container instantiations, but every read and write must cross an authorization and encryption boundary.
```

## Persistence

Persistence means data can survive after the process exits.

Process memory is not enough for state that should survive restarts because memory disappears when the process stops.

## Mounted Data Directory

A mounted data directory connects a host directory to a path inside the container.

This module mounts `data` from the module directory into `/data` inside the container.

Data survives container deletion when the host directory remains.

## Bearer Token

A bearer token is a value the client sends to prove it is allowed to use protected routes.

This module expects:

```text
Authorization: Bearer module-016-token
```

The bearer token is a trust boundary. Requests without the expected token cannot read or write records.

`AUTH_TOKEN` configures the expected token.

## Encryption Key

`ENCRYPTION_KEY` configures the bytes used to encrypt and decrypt the records file.

The bearer token is not the encryption key. The bearer token decides who may cross the HTTP boundary. The encryption key decides whether the bytes on disk can become records.

This module requires `ENCRYPTION_KEY` to be exactly 32 bytes so it can use AES-256-GCM.

The demo key is:

```text
0123456789abcdef0123456789abcdef
```

This is a demo key for this learning module, not a production secret.

## AES-GCM

AES-GCM is an authenticated encryption mode from the Go standard library.

At a beginner level, it turns plaintext records into encrypted bytes and detects when encrypted bytes cannot be opened with the key.

This module creates a fresh random nonce for every encryption.

The file format is intentionally simple:

```text
nonce || ciphertext
```

## Encrypted File

The file on disk is encrypted so the stored records do not sit in the mounted data directory as plain JSON.

The service decrypts the file, edits records in memory, and re-encrypts the file after a write.

This is not JWT, OAuth, KMS, Vault, or database persistence yet.

## Primitive Counters

This module still uses primitive counters instead of a metrics framework.

The `/metrics` route exposes authorized requests, unauthorized requests, read successes, write successes, decrypt failures, and encrypt failures.

## Earlier Modules Still Apply

Module 012 graceful shutdown still applies. The server still waits for an operating-system signal and exits cleanly.

Module 013 request lifetime and instrumentation still apply. The service still counts important outcomes and exposes those counts through `/metrics`.

Module 014 server timeouts still apply. The explicit `http.Server` timeout fields still protect the connection boundary.

Module 015 rate-limit thinking still applies as capacity protection, even though this module does not add a rate limiter.

## The Program

```go
package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

type config struct {
	Port          string
	AuthToken     string
	EncryptionKey []byte
	DataFile      string
}

type appState struct {
	config  config
	metrics *metrics
	fileMu  sync.Mutex
}

type metrics struct {
	mu                   sync.Mutex
	requestsTotal        int
	authorizedTotal      int
	unauthorizedTotal    int
	recordReadsTotal     int
	recordWritesTotal    int
	decryptFailuresTotal int
	encryptFailuresTotal int
}

type record struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type recordStore struct {
	Records map[string]record `json:"records"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type metricsResponse struct {
	RequestsTotal        int `json:"requests_total"`
	AuthorizedTotal      int `json:"authorized_total"`
	UnauthorizedTotal    int `json:"unauthorized_total"`
	RecordReadsTotal     int `json:"record_reads_total"`
	RecordWritesTotal    int `json:"record_writes_total"`
	DecryptFailuresTotal int `json:"decrypt_failures_total"`
	EncryptFailuresTotal int `json:"encrypt_failures_total"`
}

type recordsResponse struct {
	Records []record `json:"records"`
}

type saveResponse struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func loadConfig() (config, error) {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		return config{}, errors.New("AUTH_TOKEN is required")
	}

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) != 32 {
		return config{}, errors.New("ENCRYPTION_KEY must be exactly 32 bytes")
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "/data/records.enc"
	}

	return config{
		Port:          port,
		AuthToken:     authToken,
		EncryptionKey: encryptionKey,
		DataFile:      dataFile,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func logRequest(r *http.Request, status int, start time.Time) {
	log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, status, time.Since(start))
}

func incrementRequests(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.requestsTotal++
}

func incrementAuthorized(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.authorizedTotal++
}

func incrementUnauthorized(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.unauthorizedTotal++
}

func incrementRecordReads(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.recordReadsTotal++
}

func incrementRecordWrites(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.recordWritesTotal++
}

func incrementDecryptFailures(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.decryptFailuresTotal++
}

func incrementEncryptFailures(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.encryptFailuresTotal++
}

func snapshotMetrics(appMetrics *metrics) metricsResponse {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()

	return metricsResponse{
		RequestsTotal:        appMetrics.requestsTotal,
		AuthorizedTotal:      appMetrics.authorizedTotal,
		UnauthorizedTotal:    appMetrics.unauthorizedTotal,
		RecordReadsTotal:     appMetrics.recordReadsTotal,
		RecordWritesTotal:    appMetrics.recordWritesTotal,
		DecryptFailuresTotal: appMetrics.decryptFailuresTotal,
		EncryptFailuresTotal: appMetrics.encryptFailuresTotal,
	}
}

func (app *appState) authorize(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != "Bearer "+app.config.AuthToken {
		incrementUnauthorized(app.metrics)
		w.Header().Set("WWW-Authenticate", "Bearer")
		log.Printf("event=auth_failed path=%s", r.URL.Path)
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return false
	}

	incrementAuthorized(app.metrics)
	return true
}

func (app *appState) loadStore() (recordStore, error) {
	encrypted, err := os.ReadFile(app.config.DataFile)
	if errors.Is(err, os.ErrNotExist) {
		return recordStore{Records: map[string]record{}}, nil
	}
	if err != nil {
		return recordStore{}, err
	}

	block, err := aes.NewCipher(app.config.EncryptionKey)
	if err != nil {
		return recordStore{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return recordStore{}, err
	}
	if len(encrypted) < gcm.NonceSize() {
		return recordStore{}, errors.New("encrypted file is too short")
	}

	nonce := encrypted[:gcm.NonceSize()]
	ciphertext := encrypted[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return recordStore{}, err
	}

	var store recordStore
	if err := json.Unmarshal(plaintext, &store); err != nil {
		return recordStore{}, err
	}
	if store.Records == nil {
		store.Records = map[string]record{}
	}

	return store, nil
}

func (app *appState) saveStore(store recordStore) error {
	plaintext, err := json.Marshal(store)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(app.config.EncryptionKey)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	encrypted := append(nonce, ciphertext...)

	if err := os.MkdirAll(filepath.Dir(app.config.DataFile), 0700); err != nil {
		return err
	}
	return os.WriteFile(app.config.DataFile, encrypted, 0600)
}

func handleHealthz(app *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(app.metrics)
		writeJSON(w, status, healthResponse{Status: "ok"})
		logRequest(r, status, start)
	}
}

func handleMetrics(app *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(app.metrics)
		writeJSON(w, status, snapshotMetrics(app.metrics))
		logRequest(r, status, start)
	}
}

func handleRecords(app *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		incrementRequests(app.metrics)

		if !app.authorize(w, r) {
			logRequest(r, http.StatusUnauthorized, start)
			return
		}

		switch r.Method {
		case http.MethodGet:
			app.handleListRecords(w, r, start)
		case http.MethodPost:
			app.handleSaveRecord(w, r, start)
		default:
			status := http.StatusMethodNotAllowed
			writeJSON(w, status, errorResponse{Error: "method not allowed"})
			logRequest(r, status, start)
		}
	}
}

func handleRecordByID(app *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		incrementRequests(app.metrics)

		if !app.authorize(w, r) {
			logRequest(r, http.StatusUnauthorized, start)
			return
		}
		if r.Method != http.MethodGet {
			status := http.StatusMethodNotAllowed
			writeJSON(w, status, errorResponse{Error: "method not allowed"})
			logRequest(r, status, start)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/records/")
		if id == "" {
			status := http.StatusNotFound
			writeJSON(w, status, errorResponse{Error: "record not found"})
			logRequest(r, status, start)
			return
		}

		app.fileMu.Lock()
		store, err := app.loadStore()
		app.fileMu.Unlock()
		if err != nil {
			status := http.StatusInternalServerError
			incrementDecryptFailures(app.metrics)
			log.Printf("event=decrypt_failed path=%s error=%v", r.URL.Path, err)
			writeJSON(w, status, errorResponse{Error: "decrypt failed"})
			logRequest(r, status, start)
			return
		}

		found, ok := store.Records[id]
		if !ok {
			status := http.StatusNotFound
			writeJSON(w, status, errorResponse{Error: "record not found"})
			logRequest(r, status, start)
			return
		}

		status := http.StatusOK
		incrementRecordReads(app.metrics)
		log.Printf("event=record_read path=%s id=%s", r.URL.Path, id)
		writeJSON(w, status, found)
		logRequest(r, status, start)
	}
}

func (app *appState) handleListRecords(w http.ResponseWriter, r *http.Request, start time.Time) {
	app.fileMu.Lock()
	store, err := app.loadStore()
	app.fileMu.Unlock()
	if err != nil {
		status := http.StatusInternalServerError
		incrementDecryptFailures(app.metrics)
		log.Printf("event=decrypt_failed path=%s error=%v", r.URL.Path, err)
		writeJSON(w, status, errorResponse{Error: "decrypt failed"})
		logRequest(r, status, start)
		return
	}

	records := make([]record, 0, len(store.Records))
	for _, item := range store.Records {
		records = append(records, item)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})

	status := http.StatusOK
	incrementRecordReads(app.metrics)
	log.Printf("event=records_read path=/records count=%d", len(records))
	writeJSON(w, status, recordsResponse{Records: records})
	logRequest(r, status, start)
}

func (app *appState) handleSaveRecord(w http.ResponseWriter, r *http.Request, start time.Time) {
	var input record
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		status := http.StatusBadRequest
		writeJSON(w, status, errorResponse{Error: "invalid json"})
		logRequest(r, status, start)
		return
	}
	if input.ID == "" {
		status := http.StatusBadRequest
		writeJSON(w, status, errorResponse{Error: "missing id"})
		logRequest(r, status, start)
		return
	}

	app.fileMu.Lock()
	store, err := app.loadStore()
	if err != nil {
		app.fileMu.Unlock()
		status := http.StatusInternalServerError
		incrementDecryptFailures(app.metrics)
		log.Printf("event=decrypt_failed path=%s error=%v", r.URL.Path, err)
		writeJSON(w, status, errorResponse{Error: "decrypt failed"})
		logRequest(r, status, start)
		return
	}

	store.Records[input.ID] = input
	if err := app.saveStore(store); err != nil {
		app.fileMu.Unlock()
		status := http.StatusInternalServerError
		incrementEncryptFailures(app.metrics)
		log.Printf("event=encrypt_failed path=%s error=%v", r.URL.Path, err)
		writeJSON(w, status, errorResponse{Error: "encrypt failed"})
		logRequest(r, status, start)
		return
	}
	app.fileMu.Unlock()

	status := http.StatusOK
	incrementRecordWrites(app.metrics)
	log.Printf("event=record_written path=/records id=%s", input.ID)
	writeJSON(w, status, saveResponse{Status: "saved", ID: input.ID})
	logRequest(r, status, start)
}

func main() {
	appConfig, err := loadConfig()
	if err != nil {
		log.Printf("event=config_error error=%v", err)
		os.Exit(1)
	}

	app := &appState{
		config:  appConfig,
		metrics: &metrics{},
	}

	http.HandleFunc("/healthz", handleHealthz(app))
	http.HandleFunc("/metrics", handleMetrics(app))
	http.HandleFunc("/records", handleRecords(app))
	http.HandleFunc("/records/", handleRecordByID(app))

	address := ":" + app.config.Port
	server := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/metrics,/records,/records/{id}", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-shutdownSignal.Done():
		log.Printf("event=shutdown_start")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("event=shutdown_failed error=%v", err)
			return
		}

		if err := <-serverErr; err != nil {
			log.Printf("event=server_failed error=%v", err)
			return
		}

		log.Printf("event=shutdown_complete")
	case err := <-serverErr:
		if err != nil {
			log.Printf("event=server_failed error=%v", err)
		}
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-016 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
mkdir -p data
docker rm -f go-scaling-module-016 >/dev/null 2>&1 || true
docker run --rm -d \
  --name go-scaling-module-016 \
  -p 8080:8080 \
  -e APP_PORT=8080 \
  -e AUTH_TOKEN=module-016-token \
  -e ENCRYPTION_KEY=0123456789abcdef0123456789abcdef \
  -e DATA_FILE=/data/records.enc \
  -v "$PWD/data:/data" \
  go-scaling:module-016
docker logs -f go-scaling-module-016 &
LOG_PID=$!
```

The server is now running in the background, the log follower keeps container output visible in the same terminal, and the `data` directory is mounted into the container.

From the same terminal, call the public routes:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
```

Make an unauthorized read:

```bash
curl -i http://localhost:8080/records
```

This should return `401 Unauthorized`.

Write an authorized record:

```bash
curl -i -X POST http://localhost:8080/records \
  -H 'Authorization: Bearer module-016-token' \
  -H 'Content-Type: application/json' \
  -d '{"id":"alpha","text":"saved across container instantiations"}'
```

Read authorized records:

```bash
curl -i http://localhost:8080/records \
  -H 'Authorization: Bearer module-016-token'

curl -i http://localhost:8080/records/alpha \
  -H 'Authorization: Bearer module-016-token'

curl http://localhost:8080/metrics
```

Show that the host file exists but is encrypted bytes, not plain JSON:

```bash
ls -l data
cat data/records.enc
```

`cat` may show unreadable output because the file is encrypted.

Stop the first container:

```bash
docker stop go-scaling-module-016
wait "$LOG_PID" 2>/dev/null || true
```

Start a second container using the same mounted `data` directory:

```bash
docker rm -f go-scaling-module-016-second >/dev/null 2>&1 || true
docker run --rm -d \
  --name go-scaling-module-016-second \
  -p 8080:8080 \
  -e APP_PORT=8080 \
  -e AUTH_TOKEN=module-016-token \
  -e ENCRYPTION_KEY=0123456789abcdef0123456789abcdef \
  -e DATA_FILE=/data/records.enc \
  -v "$PWD/data:/data" \
  go-scaling:module-016
docker logs -f go-scaling-module-016-second &
LOG_PID=$!
```

Read the same record from the second container:

```bash
curl -i http://localhost:8080/records/alpha \
  -H 'Authorization: Bearer module-016-token'
```

The record survived because the encrypted file lived in the mounted host directory, not only in the deleted container.

When you are done, stop the second container deterministically by name:

```bash
docker stop go-scaling-module-016-second
wait "$LOG_PID" 2>/dev/null || true
```

## What Changed From Module 015

Module 015 added admission control before work begins.

This module keeps the service boundary, counters, timeouts, and graceful shutdown. It adds bearer-token authorization and one encrypted JSON file that survives container instantiations through a mounted data directory.

## First-Principles Chain

```text
source code → server boundary → request → bearer token → authorization decision → encrypted file → decrypt → edit records → encrypt → persisted bytes → metrics response → logs → graceful shutdown → process exit
```

Source code receives a request at the server boundary. Protected requests must provide the bearer token. Authorized reads and writes decrypt the encrypted file, work with records in memory, and write encrypted bytes back to disk. Metrics and logs expose outcomes, and graceful shutdown still controls process exit.

## Why Bearer-Token Encrypted Persistence Matters For Scaling Later

Larger services need state that can survive restarts.

Sensitive state needs boundaries. The bearer token controls access to the route, and AES-GCM keeps the file from being plain JSON on disk.
