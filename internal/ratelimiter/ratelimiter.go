package ratelimiter

import (
	"sync"
	"time"
)

// BucketConfig содержит настройки bucket для конкретного клиента или endpoint'а
type BucketConfig struct {
	Capacity int
	Rate     int
	Custom   bool // Показывает, является ли конфиг кастомным для клиента
}

// TokenBucket реализация с оптимизированной блокировкой
type TokenBucket struct {
	config     BucketConfig
	tokens     int32
	lastRefill int64 // UnixNano
	mu         sync.Mutex
}

// NewTokenBucket создает новый bucket с заданной конфигурацией
func NewTokenBucket(config BucketConfig) *TokenBucket {
	return &TokenBucket{
		config:     config,
		tokens:     int32(config.Capacity),
		lastRefill: time.Now().UnixNano(),
	}
}

// Take пытается взять токен из bucket
func (tb *TokenBucket) Take(now time.Time) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Пополняем bucket
	nsSinceRefill := now.UnixNano() - tb.lastRefill
	secondsSinceRefill := float64(nsSinceRefill) / 1e9
	tokensToAdd := int32(secondsSinceRefill * float64(tb.config.Rate))

	if tokensToAdd > 0 {
		if newTokens := tb.tokens + tokensToAdd; newTokens <= int32(tb.config.Capacity) {
			tb.tokens = newTokens
		} else {
			tb.tokens = int32(tb.config.Capacity)
		}
		tb.lastRefill = now.UnixNano()
	}

	// Проверяем доступность токена
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// Client представляет клиента с набором bucket'ов
type Client struct {
	ip       string
	buckets  sync.Map // map[string]*TokenBucket
	lastSeen int64    // atomic
}

// RateLimiter основной тип для ограничения запросов
type RateLimiter struct {
	defaultConfig BucketConfig
	clients       sync.Map // map[string]*Client
	db            ConfigDB
	closeCh       chan struct{}
}

// ConfigDB интерфейс для хранения кастомных конфигураций
type ConfigDB interface {
	GetConfig(clientIP, endpoint string) (BucketConfig, error)
}

// NewRateLimiter создает новый RateLimiter
func NewRateLimiter(defaultCapacity, defaultRate int, db ConfigDB) *RateLimiter {
	rl := &RateLimiter{
		defaultConfig: BucketConfig{
			Capacity: defaultCapacity,
			Rate:     defaultRate,
		},
		db:      db,
		closeCh: make(chan struct{}),
	}

	go rl.cleanupInactiveClients()
	return rl
}

// Allow проверяет, разрешен ли запрос
func (rl *RateLimiter) Allow(clientIP, endpoint string) bool {
	now := time.Now()

	// Получаем или создаем клиента
	client, _ := rl.clients.LoadOrStore(clientIP, &Client{
		ip:       clientIP,
		lastSeen: now.UnixNano(),
	})

	// Обновляем время последней активности
	client.(*Client).lastSeen = now.UnixNano()

	// Получаем или создаем bucket
	bucket, _ := client.(*Client).buckets.LoadOrStore(endpoint, func() *TokenBucket {
		// Пытаемся получить кастомную конфигурацию
		if rl.db != nil {
			if config, err := rl.db.GetConfig(clientIP, endpoint); err == nil && config.Custom {
				return NewTokenBucket(config)
			}
		}
		return NewTokenBucket(rl.defaultConfig)
	}())

	return bucket.(*TokenBucket).Take(now)
}

// Stop останавливает фоновые процессы
func (rl *RateLimiter) Stop() {
	close(rl.closeCh)
}

// cleanupInactiveClients периодически чистит неактивных клиентов
func (rl *RateLimiter) cleanupInactiveClients() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			threshold := time.Now().Add(-24 * time.Hour).UnixNano()
			rl.clients.Range(func(key, value interface{}) bool {
				if value.(*Client).lastSeen < threshold {
					rl.clients.Delete(key)
				}
				return true
			})
		case <-rl.closeCh:
			return
		}
	}
}
