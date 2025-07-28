package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"2fa-system/config"
)

type SendCodeRequest struct {
	Email string `json:"email"`
}

type SendCodeResponse struct {
	Message string `json:"message"`
}

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

func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendCodeHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SendCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Requisição inválida", http.StatusBadRequest)
			return
		}

		if !rateLimit(req.Email) {
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
			return
		}

		ip := r.RemoteAddr
		if !ipRateLimit(ip) {
			http.Error(w, "Too many requests from your IP, please try again later.", http.StatusTooManyRequests)
			return
		}

		code := generateCode()

		subject := "Seu código 2FA"
		body := fmt.Sprintf("Seu código de verificação é: %s", code)
		msg := "From: " + cfg.FromEmail + "\n" +
			"To: " + req.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			body

		auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
		err := smtp.SendMail(
			cfg.SMTPHost+":"+cfg.SMTPPort,
			auth,
			cfg.FromEmail,
			[]string{req.Email},
			[]byte(msg),
		)
		if err != nil {
			log.Printf("Erro ao enviar e-mail: %v", err)
			http.Error(w, "Erro ao enviar e-mail", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SendCodeResponse{Message: "Código enviado com sucesso"})
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
