package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// Code management
func (r *RedisStore) SaveCode(email, code string) error {
	key := fmt.Sprintf("code:%s", email)
	return r.client.Set(r.ctx, key, code, 10*time.Minute).Err() // 10 min TTL
}

func (r *RedisStore) GetCode(email string) (string, bool, error) {
	key := fmt.Sprintf("code:%s", email)
	code, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return code, true, nil
}

func (r *RedisStore) DeleteCode(email string) error {
	key := fmt.Sprintf("code:%s", email)
	return r.client.Del(r.ctx, key).Err()
}

// Rate limiting
func (r *RedisStore) SaveRateLimit(key string, duration time.Duration) error {
	rateKey := fmt.Sprintf("rate:%s", key)
	return r.client.Set(r.ctx, rateKey, time.Now().Add(duration), duration).Err()
}

func (r *RedisStore) IsRateLimited(key string) (bool, error) {
	rateKey := fmt.Sprintf("rate:%s", key)
	_, err := r.client.Get(r.ctx, rateKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Brute force protection
func (r *RedisStore) IncrementFailedAttempts(email string) (int, error) {
	key := fmt.Sprintf("failed:%s", email)
	attempts, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	
	// Set TTL for failed attempts (24 hours)
	r.client.Expire(r.ctx, key, 24*time.Hour)
	
	return int(attempts), nil
}

func (r *RedisStore) ResetFailedAttempts(email string) error {
	key := fmt.Sprintf("failed:%s", email)
	blockKey := fmt.Sprintf("blocked:%s", email)
	
	// Delete both failed attempts and blocked status
	r.client.Del(r.ctx, key, blockKey)
	return nil
}

func (r *RedisStore) IsBlocked(email string) (bool, time.Time, error) {
	key := fmt.Sprintf("blocked:%s", email)
	untilStr, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, time.Time{}, err
	}
	
	var until time.Time
	if err := json.Unmarshal([]byte(untilStr), &until); err != nil {
		return false, time.Time{}, err
	}
	
	if time.Now().Before(until) {
		return true, until, nil
	}
	
	// Remove expired block
	r.client.Del(r.ctx, key)
	return false, time.Time{}, nil
}

func (r *RedisStore) SetBlocked(email string, until time.Time) error {
	key := fmt.Sprintf("blocked:%s", email)
	untilBytes, err := json.Marshal(until)
	if err != nil {
		return err
	}
	
	duration := time.Until(until)
	if duration < 0 {
		duration = 1 * time.Hour // Default TTL
	}
	
	return r.client.Set(r.ctx, key, untilBytes, duration).Err()
}

// Monitoring
func (r *RedisStore) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get all rate limit keys
	ratePattern := "rate:*"
	rateKeys, err := r.client.Keys(r.ctx, ratePattern).Result()
	if err != nil {
		return nil, err
	}
	
	emailRateLimit := make(map[string]interface{})
	ipRateLimit := make(map[string]interface{})
	
	for _, key := range rateKeys {
		// Extract the actual key (remove "rate:" prefix)
		actualKey := key[5:] // Remove "rate:" prefix
		
		// Get the expiration time
		ttl, err := r.client.TTL(r.ctx, key).Result()
		if err != nil {
			continue
		}
		
		if ttl > 0 {
			blockedUntil := time.Now().Add(ttl)
			
			if len(actualKey) > 3 && actualKey[:3] == "ip:" {
				// IP rate limit
				ip := actualKey[3:]
				ipRateLimit[ip] = map[string]interface{}{
					"blocked":     true,
					"blockedUntil": blockedUntil,
				}
			} else {
				// Email rate limit
				emailRateLimit[actualKey] = map[string]interface{}{
					"blocked":     true,
					"blockedUntil": blockedUntil,
				}
			}
		}
	}
	
	stats["emailRateLimit"] = emailRateLimit
	stats["ipRateLimit"] = ipRateLimit
	
	// Get failed attempts
	failedPattern := "failed:*"
	failedKeys, err := r.client.Keys(r.ctx, failedPattern).Result()
	if err != nil {
		return nil, err
	}
	
	failedAttempts := make(map[string]interface{})
	for _, key := range failedKeys {
		email := key[7:] // Remove "failed:" prefix
		attempts, err := r.client.Get(r.ctx, key).Int()
		if err != nil {
			continue
		}
		
		if attempts > 0 {
			failedAttempts[email] = map[string]interface{}{
				"attempts": attempts,
			}
		}
	}
	stats["failedAttempts"] = failedAttempts
	
	// Get blocked users
	blockedPattern := "blocked:*"
	blockedKeys, err := r.client.Keys(r.ctx, blockedPattern).Result()
	if err != nil {
		return nil, err
	}
	
	blockedUntil := make(map[string]interface{})
	for _, key := range blockedKeys {
		email := key[8:] // Remove "blocked:" prefix
		untilStr, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue
		}
		
		var until time.Time
		if err := json.Unmarshal([]byte(untilStr), &until); err != nil {
			continue
		}
		
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