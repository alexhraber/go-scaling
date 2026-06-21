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
