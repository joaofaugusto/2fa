package services

import (
	"fmt"
	"math/rand"
	"time"

	"2fa-system/storage"
)

type CodeService struct {
	Store storage.Storage
}

func NewCodeService(store storage.Storage) *CodeService {
	return &CodeService{Store: store}
}

func (s *CodeService) GenerateAndSaveCode(email string) string {
	code := generateCode()
	s.Store.SaveCode(email, code)
	return code
}

func (s *CodeService) ValidateCode(email, code string) bool {
	storedCode, exists, err := s.Store.GetCode(email)
	if err != nil || !exists || storedCode != code {
		return false
	}
	s.Store.DeleteCode(email) // Remove após uso
	return true
}

func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
