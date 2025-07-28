package handlers

import (
	"encoding/json"
	"net/http"

	"2fa-system/models"
	"2fa-system/services"
	"2fa-system/storage"
)

func VerifyCodeHandler(codeService *services.CodeService, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req models.VerifyCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Requisição inválida", http.StatusBadRequest)
			return
		}
		// Reset failed attempts on success
		store.ResetFailedAttempts(req.Email)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.VerifyCodeResponse{
			Success: true,
			Message: "Código verificado com sucesso",
		})
	}
}

func MonitorHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		
		// Serve the HTML template
		http.ServeFile(w, r, "templates/monitor.html")
	}
}

func MonitorDataHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		stats, err := store.GetStats()
		if err != nil {
			http.Error(w, "Error getting stats", http.StatusInternalServerError)
			return
		}
		
		json.NewEncoder(w).Encode(stats)
	}
}
