package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/amayones/authz-engine/internal/api"
	"github.com/amayones/authz-engine/internal/migrate"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
	"github.com/amayones/authz-engine/internal/store/mssql"
	"github.com/amayones/authz-engine/internal/store/postgres"
)

func main() {
	connString := flag.String("db", "", "Database connection string")
	clientName := flag.String("name", "", "Nama client (misal: 'billing-service')")
	rpm := flag.Int("rpm", 60, "Rate limit request per menit")
	driver := flag.String("driver", "sqlserver", "Database driver: 'sqlserver' atau 'postgres'")
	flag.Parse()

	if *connString == "" || *clientName == "" {
		log.Fatal("wajib isi -db dan -name")
	}

	rawKey, hash, err := api.GenerateAndHashKey()
	if err != nil {
		log.Fatalf("gagal generate key: %v", err)
	}

	var s store.Store
	switch migrate.Driver(*driver) {
	case migrate.DriverPostgres:
		pg, err := postgres.New(*connString)
		if err != nil {
			log.Fatalf("gagal konek database postgres: %v", err)
		}
		defer pg.Close()
		s = pg
	default:
		ms, err := mssql.New(*connString)
		if err != nil {
			log.Fatalf("gagal konek database sqlserver: %v", err)
		}
		defer ms.Close()
		s = ms
	}

	err = s.CreateAPIKey(context.Background(), hash, model.APIKey{
		ClientName:   *clientName,
		RateLimitRPM: *rpm,
		IsActive:     true,
	})
	if err != nil {
		log.Fatalf("gagal simpan api key: %v", err)
	}

	fmt.Println("API key berhasil dibuat. SIMPAN SEKARANG — tidak akan ditampilkan lagi:")
	fmt.Println(rawKey)
}
