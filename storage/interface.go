package storage

import "time"

// Storage interface for different storage backends
type Storage interface {
	// Code management
	SaveCode(email, code string) error
	GetCode(email string) (string, bool, error)
	DeleteCode(email string) error
	
	// Rate limiting
	SaveRateLimit(key string, duration time.Duration) error
	IsRateLimited(key string) (bool, error)
	
	// Brute force protection
	IncrementFailedAttempts(email string) (int, error)
	ResetFailedAttempts(email string) error
	IsBlocked(email string) (bool, time.Time, error)
	SetBlocked(email string, until time.Time) error
	
	// Monitoring
	GetStats() (map[string]interface{}, error)
} 