package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Simulação de armazenamento em memória dos códigos enviados
var (
	codes      = make(map[string]string) // email -> code
	codesMutex sync.RWMutex
)

type VerifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Função para salvar código (chame no SendCodeHandler)
func SaveCode(email, code string) {
	codesMutex.Lock()
	defer codesMutex.Unlock()
	codes[email] = code
}

// Função para verificar código
func VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	codesMutex.RLock()
	storedCode, exists := codes[req.Email]
	codesMutex.RUnlock()
	

	if !exists || storedCode != req.Code {
		json.NewEncoder(w).Encode(VerifyCodeResponse{
			Success: false,
			Message: "Código inválido ou expirado",
		})
		return
	}
	codesMutex.Lock()
	delete(codes, req.Email)
	codesMutex.Unlock()

	json.NewEncoder(w).Encode(VerifyCodeResponse{
		Success: true,
		Message: "Código verificado com sucesso",
	})
}
