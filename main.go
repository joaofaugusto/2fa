package main

import (
	"log"
	"net/http"

	"2fa-system/config"
	"2fa-system/handlers"
	"2fa-system/services"
	"2fa-system/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Carrega configurações
	cfg := config.LoadConfig()

	// Instancia o armazenamento (memory ou Redis)
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Instancia os serviços
	codeService := services.NewCodeService(store)
	emailService := services.NewEmailService(cfg)

	// Handler para envio do código
	http.HandleFunc("/send-code", handlers.SendCodeHandler(cfg, codeService, emailService, store))

	// Handler para verificação do código
	http.HandleFunc("/verify-code", handlers.VerifyCodeHandler(codeService, store))

	// Monitor endpoints
	http.HandleFunc("/monitor", handlers.MonitorHandler(store))
	http.HandleFunc("/monitor-data", handlers.MonitorDataHandler(store))

	log.Println("Servidor iniciado na porta :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
