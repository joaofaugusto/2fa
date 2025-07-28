package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"2fa-system/models"
	"2fa-system/services"
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

var (
	failedAttempts   = make(map[string]int)
	blockedUntil     = make(map[string]time.Time)
	bruteForceMu     sync.Mutex
	maxAttempts      = 5
	blockDuration    = 10 * time.Minute
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

func checkBruteForce(email string) bool {
	bruteForceMu.Lock()
	defer bruteForceMu.Unlock()
	if until, blocked := blockedUntil[email]; blocked && time.Now().Before(until) {
		return false // Blocked
	}
	return true
}

func registerFailedAttempt(email string) {
	bruteForceMu.Lock()
	defer bruteForceMu.Unlock()
	failedAttempts[email]++
	if failedAttempts[email] >= maxAttempts {
		blockedUntil[email] = time.Now().Add(blockDuration)
		failedAttempts[email] = 0 // reset counter
	}
}

func resetFailedAttempts(email string) {
	bruteForceMu.Lock()
	defer bruteForceMu.Unlock()
	failedAttempts[email] = 0
	delete(blockedUntil, email)
}

func VerifyCodeHandler(codeService *services.CodeService) http.HandlerFunc {
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

		ip := r.RemoteAddr
		if !ipRateLimit(ip) {
			http.Error(w, "Too many requests from your IP, please try again later.", http.StatusTooManyRequests)
			return
		}

		if !rateLimit(req.Email) {
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
			return
		}

		if !checkBruteForce(req.Email) {
			http.Error(w, "Too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		if !codeService.ValidateCode(req.Email, req.Code) {
			registerFailedAttempt(req.Email)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.VerifyCodeResponse{
				Success: false,
				Message: "Código inválido ou expirado",
			})
		} else {
			resetFailedAttempts(req.Email)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.VerifyCodeResponse{
				Success: true,
				Message: "Código verificado com sucesso",
			})
		}
	}
}

func MonitorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>2FA Monitor Panel</h1>")
		fmt.Fprintf(w, "<h2>Per-Email Rate Limit</h2><pre>%v</pre>", rateLimitMap)
		fmt.Fprintf(w, "<h2>Per-IP Rate Limit</h2><pre>%v</pre>", ipRateLimitMap)
		fmt.Fprintf(w, "<h2>Failed Attempts</h2><pre>%v</pre>", failedAttempts)
		fmt.Fprintf(w, "<h2>Blocked Until</h2><pre>%v</pre>", blockedUntil)
	}
}
