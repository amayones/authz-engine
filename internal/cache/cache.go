// Package cache berisi implementasi TTL cache sederhana, thread-safe,
// dipakai untuk menyimpan hasil evaluasi authorization sementara.
package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     bool
	expiresAt time.Time
}

// Cache adalah in-memory TTL cache untuk boolean authorization decisions.
type Cache struct {
	mu   sync.RWMutex
	data map[string]entry
	ttl  time.Duration
}

// New membuat cache baru dengan TTL tertentu (misal 10*time.Second).
func New(ttl time.Duration) *Cache {
	c := &Cache{
		data: make(map[string]entry),
		ttl:  ttl,
	}
	go c.janitor() // background cleanup entry yang expired
	return c
}

func (c *Cache) Get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, exists := c.data[key]
	if !exists || time.Now().After(e.expiresAt) {
		return false, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

// Invalidate menghapus satu key spesifik.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// InvalidatePrefix menghapus semua key yang diawali prefix tertentu.
// Dipakai kalau kita tidak tahu key persis (misal semua cache milik
// satu subject atau satu object).
func (c *Cache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.data, k)
		}
	}
}

// Clear mengosongkan seluruh cache. Dipakai sebagai fallback paling aman
// kalau ragu apa yang perlu di-invalidate.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]entry)
}

// janitor membersihkan entry expired setiap beberapa detik supaya map
// tidak membengkak terus walau tidak pernah di-Get lagi.
func (c *Cache) janitor() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.data {
			if now.After(e.expiresAt) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}
