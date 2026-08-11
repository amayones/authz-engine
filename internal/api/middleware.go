package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/store"
)

// loggingMiddleware mencatat tiap request: method, path, status, durasi.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

// statusWriter membungkus ResponseWriter supaya kita bisa tahu status
// code yang akhirnya dikirim (net/http tidak expose ini secara default).
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

// recoveryMiddleware mencegah satu panic di satu request menjatuhkan
// seluruh server — penting karena ini authorization service, harus
// tetap hidup untuk melayani request lain walau satu request error fatal.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// clientContextKey dipakai untuk menyimpan info client (dari API key)
// ke request context, supaya handler bisa tahu siapa yang memanggil
// kalau perlu (misal untuk audit log nanti).
type contextKey string

const clientContextKey contextKey = "client"

// authMiddleware memverifikasi API key terhadap database, lalu terapkan
// rate limit sesuai RateLimitRPM milik key tersebut.
func authMiddleware(s store.APIKeyStore, limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			rawKey := r.Header.Get("X-API-Key")
			if rawKey == "" {
				writeError(w, http.StatusUnauthorized, "missing X-API-Key header")
				return
			}

			hash := hashAPIKey(rawKey)
			apiKey, err := s.GetAPIKeyByHash(r.Context(), hash)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			if !apiKey.IsActive {
				writeError(w, http.StatusUnauthorized, "API key has been revoked")
				return
			}

			if !limiter.Allow(hash, apiKey.RateLimitRPM) {
				rateLimitRejectedTotal.Inc()
				w.Header().Set("Retry-After", "60")
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			ctx := engine.WithActor(r.Context(), apiKey.ClientName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// chain menggabungkan beberapa middleware jadi satu, dieksekusi
// berurutan sesuai urutan parameter.
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// timeoutMiddleware memaksa tiap request punya batas waktu maksimum.
// Ini mencegah satu request yang macet (misal query database yang
// lambat) menahan resource selamanya.
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, `{"error":"request timeout"}`)
	}
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(sw.status)
		requestsTotal.WithLabelValues(r.URL.Path, r.Method, status).Inc()
		requestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}
