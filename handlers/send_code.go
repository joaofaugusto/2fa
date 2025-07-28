package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"2fa-system/config"
	"2fa-system/models"
	"2fa-system/services"
	"2fa-system/storage"
)

func SendCodeHandler(cfg *config.Config, codeService *services.CodeService, emailService *services.EmailService, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req models.SendCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Requisição inválida", http.StatusBadRequest)
			return
		}

		// Check IP rate limiting
		ip := r.RemoteAddr
		ipKey := "ip:" + ip
		if isLimited, err := store.IsRateLimited(ipKey); err == nil && isLimited {
			http.Error(w, "Too many requests from your IP, please try again later.", http.StatusTooManyRequests)
			return
		}

		// Check email rate limiting
		if isLimited, err := store.IsRateLimited(req.Email); err == nil && isLimited {
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
			return
		}

		// Set rate limits
		store.SaveRateLimit(ipKey, time.Minute)
		store.SaveRateLimit(req.Email, time.Minute)

		code := codeService.GenerateAndSaveCode(req.Email)
		if err := emailService.SendCodeEmail(req.Email, code); err != nil {
			log.Printf("Erro ao enviar e-mail: %v", err)
			http.Error(w, "Erro ao enviar e-mail", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.SendCodeResponse{Message: "Código enviado com sucesso"})
	}
}
