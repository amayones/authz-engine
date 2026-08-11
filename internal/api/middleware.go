package api

import (
	"log"
	"net/http"
	"time"
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

// apiKeyMiddleware adalah proteksi paling dasar: cek header X-API-Key.
// Ini BUKAN pengganti auth production-grade (belum ada rotation, rate
// limit per key, dll) — cukup untuk mencegah akses publik tanpa izin
// selama development/internal use.
func apiKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health check tidak perlu API key, supaya load balancer bisa cek.
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" || key != expectedKey {
				writeError(w, http.StatusUnauthorized, "missing or invalid API key")
				return
			}
			next.ServeHTTP(w, r)
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
