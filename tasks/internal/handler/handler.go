package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/omnikk/pz4.2/tasks/internal/metrics"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	log      *logrus.Entry
	authAddr string
}

func New(log *logrus.Entry, authAddr string) *Handler {
	return &Handler{log: log, authAddr: authAddr}
}

type Task struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (h *Handler) wrap(route string, next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.InFlightRequests.Inc()
		defer metrics.InFlightRequests.Dec()

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.status)

		metrics.RequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, route).Observe(duration)
	}
}

func (h *Handler) verifyToken(r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		return false
	}
	req, err := http.NewRequestWithContext(r.Context(), "GET",
		fmt.Sprintf("%s/v1/auth/verify", h.authAddr), nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("X-Request-ID", r.Header.Get("X-Request-ID"))

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if !h.verifyToken(r) {
		h.log.Warn("unauthorized request to list tasks")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	tasks := []Task{
		{Title: "Implement metrics", Description: "Add prometheus to tasks", DueDate: "2026-01-12"},
		{Title: "Write tests", Description: "Cover handlers with tests", DueDate: "2026-01-15"},
	}
	h.log.Info("tasks listed successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if !h.verifyToken(r) {
		h.log.Warn("unauthorized request to create task")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.log.WithField("error", err.Error()).Error("failed to decode request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	h.log.WithField("title", task.Title).Info("task created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/tasks", h.wrap("/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListTasks(w, r)
		case http.MethodPost:
			h.CreateTask(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}
