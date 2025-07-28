package storage

import (
	"sync"
	"time"
)

type MemoryStore struct {
	codes map[string]string
	mu    sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		codes: make(map[string]string),
	}
}

func (m *MemoryStore) SaveCode(email, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[email] = code
}

func (m *MemoryStore) GetCode(email string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	code, exists := m.codes[email]
	return code, exists
}

func (m *MemoryStore) DeleteCode(email string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.codes, email)
}

var (
	ipRateLimitMap = make(map[string]time.Time)
	ipRateLimitMu  sync.Mutex
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
