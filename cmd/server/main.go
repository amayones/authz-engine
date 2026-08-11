package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/amayones/authz-engine/internal/api"
	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/mssql"
)

func main() {
	connString := getEnv("AUTHZ_DB_CONN", "sqlserver://may:may@localhost:1433?database=authzdb")
	apiKey := getEnv("AUTHZ_API_KEY", "")
	addr := getEnv("AUTHZ_ADDR", ":8080")

	if apiKey == "" {
		log.Fatal("AUTHZ_API_KEY wajib diisi, jangan jalankan server tanpa proteksi")
	}

	st, err := mssql.New(connString)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	defer st.Close()

	e := engine.NewWithCache(st, 15*time.Second)

	// Schema hierarki relasi didefinisikan sekali di startup.
	// Sesuaikan dengan kebutuhan aplikasi kamu.
	e.SetSchema(model.RelationSchema{
		"viewer": {"editor", "owner"},
		"editor": {"owner"},
	})

	server := api.NewServer(e)
	router := api.NewRouter(server, apiKey)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("authz-engine listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
