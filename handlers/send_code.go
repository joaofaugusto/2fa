package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"2fa-system/config"
	"2fa-system/models"
	"2fa-system/services"
	"2fa-system/storage"
)

var (
	rateLimitMap = make(map[string]time.Time)
	rateLimitMu  sync.Mutex
	rateLimitDuration = time.Minute // 1 request per minute per email
)

var (
	ipRateLimitMap = make(map[string]time.Time)
	ipRateLimitMu  sync.Mutex
	ipRateLimitDuration = time.Minute // 1 request per minute per IP
)

func rateLimit(email string) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()
	last, exists := rateLimitMap[email]
	if exists && time.Since(last) < rateLimitDuration {
		return false // Blocked
	}
	rateLimitMap[email] = time.Now()
	return true // Allowed
}

func ipRateLimit(ip string) bool {
	ipRateLimitMu.Lock()
	defer ipRateLimitMu.Unlock()
	last, exists := ipRateLimitMap[ip]
	if exists && time.Since(last) < ipRateLimitDuration {
		return false // Blocked
	}
	ipRateLimitMap[ip] = time.Now()
	return true // Allowed
}

func SendCodeHandler(cfg *config.Config, codeService *services.CodeService, emailService *services.EmailService) http.HandlerFunc {
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

		ip := r.RemoteAddr
		if !ipRateLimit(ip) {
			http.Error(w, "Too many requests from your IP, please try again later.", http.StatusTooManyRequests)
			return
		}

		if !rateLimit(req.Email) {
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
			return
		}

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

func MonitorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>2FA Monitor Panel</h1>")
		fmt.Fprintf(w, "<h2>Per-Email Rate Limit</h2><pre>%v</pre>", rateLimitMap)
		fmt.Fprintf(w, "<h2>Per-IP Rate Limit</h2><pre>%v</pre>", ipRateLimitMap)
	}
}
