package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nilbyte-studios/network-infra/internal/config"
	"github.com/nilbyte-studios/network-infra/internal/metrics"
	"github.com/nilbyte-studios/network-infra/internal/storage"
	"github.com/nilbyte-studios/network-infra/internal/tailscale"
)

// Handler gerencia os endpoints HTTP
type Handler struct {
	collector *metrics.Collector
	storage   *storage.Storage
	config    *config.Config
}

// NewHandler cria um novo handler
func NewHandler(collector *metrics.Collector, storage *storage.Storage, cfg *config.Config) *Handler {
	return &Handler{
		collector: collector,
		storage:   storage,
		config:    cfg,
	}
}

// HandleStatus retorna as métricas atuais do sistema
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verifica se a requisição vem da rede Tailscale
	remoteIP := getRemoteIP(r)
	if !tailscale.IsTailscaleIP(remoteIP) {
		http.Error(w, "Forbidden: Only Tailscale network allowed", http.StatusForbidden)
		return
	}

	// Coleta métricas atuais
	metrics, err := h.collector.Collect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to collect metrics: %v", err), http.StatusInternalServerError)
		return
	}

	// Salva no banco de dados
	if err := h.storage.Save(metrics); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to save metrics: %v\n", err)
	}

	// Retorna JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleHealth retorna status de saúde do serviço
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePeers retorna lista de peers Tailscale
func (h *Handler) HandlePeers(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verifica se a requisição vem da rede Tailscale
	remoteIP := getRemoteIP(r)
	if !tailscale.IsTailscaleIP(remoteIP) {
		http.Error(w, "Forbidden: Only Tailscale network allowed", http.StatusForbidden)
		return
	}

	// Busca peers usando API ou CLI
	peers, err := tailscale.GetPeersWithAPI(
		h.config.Tailscale.APIKey,
		h.config.Tailscale.Tailnet,
		h.config.Tailscale.UseCLI,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get peers: %v", err), http.StatusInternalServerError)
		return
	}

	// Retorna JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// HandleHistory retorna histórico de métricas
func (h *Handler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verifica se a requisição vem da rede Tailscale
	remoteIP := getRemoteIP(r)
	if !tailscale.IsTailscaleIP(remoteIP) {
		http.Error(w, "Forbidden: Only Tailscale network allowed", http.StatusForbidden)
		return
	}

	// Parse query parameter 'hours' (default: 24)
	hoursStr := r.URL.Query().Get("hours")
	hours := 24
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	// Busca histórico
	history, err := h.storage.GetHistory(hours)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history: %v", err), http.StatusInternalServerError)
		return
	}

	// Retorna JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hours":   hours,
		"count":   len(history),
		"metrics": history,
	})
}

// getRemoteIP extrai o IP remoto da requisição
func getRemoteIP(r *http.Request) string {
	// Tenta X-Forwarded-For primeiro
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Tenta X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Usa RemoteAddr
	ip := r.RemoteAddr
	if idx := len(ip) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if ip[i] == ':' {
				return ip[:i]
			}
		}
	}

	return ip
}
