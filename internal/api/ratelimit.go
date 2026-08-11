package api

import (
	"sync"
	"time"
)

// bucket adalah token bucket sederhana untuk satu client.
type bucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // token per detik
	lastRefill time.Time
}

func (b *bucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RateLimiter mengelola satu bucket per client (key = client identifier,
// biasanya keyHash), thread-safe untuk dipakai concurrent oleh banyak
// goroutine request.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{buckets: make(map[string]*bucket)}
	go rl.janitor()
	return rl
}

// Allow mengecek apakah client dengan identifier tertentu masih punya
// kuota, sekaligus membuat bucket baru kalau ini request pertamanya.
// rpm adalah rate limit request-per-menit khusus client ini (diambil
// dari APIKey.RateLimitRPM, bisa beda tiap client).
func (rl *RateLimiter) Allow(clientID string, rpm int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[clientID]
	if !exists {
		b = &bucket{
			tokens:     float64(rpm),
			maxTokens:  float64(rpm),
			refillRate: float64(rpm) / 60.0, // rpm -> token per detik
			lastRefill: time.Now(),
		}
		rl.buckets[clientID] = b
	}
	return b.allow()
}

// janitor membersihkan bucket yang tidak dipakai lebih dari 10 menit,
// supaya map tidak membengkak terus kalau ada banyak client yang datang
// lalu pergi (misal key yang di-revoke).
func (rl *RateLimiter) janitor() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for id, b := range rl.buckets {
			if time.Since(b.lastRefill) > 10*time.Minute {
				delete(rl.buckets, id)
			}
		}
		rl.mu.Unlock()
	}
}
