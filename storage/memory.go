package storage

import (
	"sync"
	"time"
)

type MemoryStore struct {
	codes            map[string]string
	rateLimits       map[string]time.Time
	failedAttempts   map[string]int
	blockedUntil     map[string]time.Time
	mu               sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		codes:          make(map[string]string),
		rateLimits:     make(map[string]time.Time),
		failedAttempts: make(map[string]int),
		blockedUntil:   make(map[string]time.Time),
	}
}

// Code management
func (m *MemoryStore) SaveCode(email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[email] = code
	return nil
}

func (m *MemoryStore) GetCode(email string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	code, exists := m.codes[email]
	return code, exists, nil
}

func (m *MemoryStore) DeleteCode(email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.codes, email)
	return nil
}

// Rate limiting
func (m *MemoryStore) SaveRateLimit(key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimits[key] = time.Now().Add(duration)
	return nil
}

func (m *MemoryStore) IsRateLimited(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if last, exists := m.rateLimits[key]; exists && time.Now().Before(last) {
		return true, nil
	}
	return false, nil
}

// Brute force protection
func (m *MemoryStore) IncrementFailedAttempts(email string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedAttempts[email]++
	return m.failedAttempts[email], nil
}

func (m *MemoryStore) ResetFailedAttempts(email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedAttempts[email] = 0
	delete(m.blockedUntil, email)
	return nil
}

func (m *MemoryStore) IsBlocked(email string) (bool, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if until, exists := m.blockedUntil[email]; exists && time.Now().Before(until) {
		return true, until, nil
	}
	return false, time.Time{}, nil
}

func (m *MemoryStore) SetBlocked(email string, until time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blockedUntil[email] = until
	return nil
}

// Monitoring
func (m *MemoryStore) GetStats() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]interface{})
	
	// Rate limit stats
	emailRateLimit := make(map[string]interface{})
	for email, lastTime := range m.rateLimits {
		if time.Now().Before(lastTime) {
			emailRateLimit[email] = map[string]interface{}{
				"blocked":     true,
				"blockedUntil": lastTime,
			}
		}
	}
	stats["emailRateLimit"] = emailRateLimit
	
	// IP rate limit stats (assuming IPs are stored with "ip:" prefix)
	ipRateLimit := make(map[string]interface{})
	for key, lastTime := range m.rateLimits {
		if len(key) > 3 && key[:3] == "ip:" {
			ip := key[3:]
			if time.Now().Before(lastTime) {
				ipRateLimit[ip] = map[string]interface{}{
					"blocked":     true,
					"blockedUntil": lastTime,
				}
			}
		}
	}
	stats["ipRateLimit"] = ipRateLimit
	
	// Failed attempts
	failedAttempts := make(map[string]interface{})
	for email, attempts := range m.failedAttempts {
		if attempts > 0 {
			failedAttempts[email] = map[string]interface{}{
				"attempts": attempts,
			}
		}
	}
	stats["failedAttempts"] = failedAttempts
	
	// Blocked users
	blockedUntil := make(map[string]interface{})
	for email, until := range m.blockedUntil {
		if time.Now().Before(until) {
			blockedUntil[email] = map[string]interface{}{
				"blocked":     true,
				"blockedUntil": until,
			}
		}
	}
	stats["blockedUntil"] = blockedUntil
	
	return stats, nil
}
