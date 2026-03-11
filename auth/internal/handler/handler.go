package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	log *logrus.Entry
}

func New(log *logrus.Entry) *Handler {
	return &Handler{log: log}
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	h.log.Info("token verified")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
