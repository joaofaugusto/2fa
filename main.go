package main

import (
	"log"
	"net/http"

	"2fa-system/config"
	"2fa-system/handlers"
	"2fa-system/services"
	"2fa-system/storage"
)

func main() {
	// Carrega configurações
	cfg := config.LoadConfig()

	// Instancia o armazenamento em memória
	store := storage.NewMemoryStore()

	// Instancia os serviços
	codeService := services.NewCodeService(store)
	emailService := services.NewEmailService(cfg)

	// Handler para envio do código
	http.HandleFunc("/send-code", handlers.SendCodeHandler(cfg, codeService, emailService))

	// Handler para verificação do código
	http.HandleFunc("/verify-code", handlers.VerifyCodeHandler(codeService))

	// Monitor endpoint
	http.HandleFunc("/monitor", handlers.MonitorHandler())

	log.Println("Servidor iniciado na porta :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
