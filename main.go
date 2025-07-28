package main

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

	"github.com/joho/godotenv"
)

var (
	rateLimitMap      = make(map[string]time.Time)
	rateLimitMu       sync.Mutex
	rateLimitDuration = time.Minute // 1 request per minute per email
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

var (
	failedAttempts = make(map[string]int)
	blockedUntil   = make(map[string]time.Time)
	bruteForceMu   sync.Mutex
	maxAttempts    = 5
	blockDuration  = 10 * time.Minute
)

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

var (
	ipRateLimitMap      = make(map[string]time.Time)
	ipRateLimitMu       sync.Mutex
	ipRateLimitDuration = time.Minute // 1 request per minute per IP
)

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env")
	}
	// Carrega configurações
	cfg := config.LoadConfig()

	// Instancia o armazenamento em memória
	store := storage.NewMemoryStore()

	// Instancia os serviços
	codeService := services.NewCodeService(store)
	emailService := services.NewEmailService(cfg)

	// Handler para envio do código
	http.HandleFunc("/send-code", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Handler para verificação do código
	http.HandleFunc("/verify-code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req models.VerifyCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Requisição inválida", http.StatusBadRequest)
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
	})

	// Monitor endpoint
	http.HandleFunc("/monitor", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>2FA Monitor Panel</h1>")
		fmt.Fprintf(w, "<h2>Per-Email Rate Limit</h2><pre>%v</pre>", rateLimitMap)
		fmt.Fprintf(w, "<h2>Per-IP Rate Limit</h2><pre>%v</pre>", ipRateLimitMap)
		fmt.Fprintf(w, "<h2>Failed Attempts</h2><pre>%v</pre>", failedAttempts)
		fmt.Fprintf(w, "<h2>Blocked Until</h2><pre>%v</pre>", blockedUntil)
	})

	log.Println("Servidor iniciado na porta :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
