package storage

import (
	"fmt"
	"log"
	"os"
)

// StorageType represents the type of storage to use
type StorageType string

const (
	StorageMemory StorageType = "memory"
	StorageRedis  StorageType = "redis"
)

// NewStorage creates a new storage instance based on configuration
func NewStorage() (Storage, error) {
	storageType := StorageType(os.Getenv("STORAGE_TYPE"))
	if storageType == "" {
		storageType = StorageMemory // Default to memory
	}

	switch storageType {
	case StorageMemory:
		log.Println("Using in-memory storage")
		return NewMemoryStore(), nil
		
	case StorageRedis:
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379" // Default Redis address
		}
		
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDB := 0 // Default Redis database
		
		log.Printf("Using Redis storage at %s", redisAddr)
		return NewRedisStore(redisAddr, redisPassword, redisDB)
		
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
} 