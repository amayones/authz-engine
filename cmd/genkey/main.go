package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/amayones/authz-engine/internal/api"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/mssql"
)

func main() {
	connString := flag.String("db", "", "Database connection string")
	clientName := flag.String("name", "", "Nama client (misal: 'billing-service')")
	rpm := flag.Int("rpm", 60, "Rate limit request per menit")
	flag.Parse()

	if *connString == "" || *clientName == "" {
		log.Fatal("wajib isi -db dan -name")
	}

	rawKey, hash, err := api.GenerateAndHashKey()
	if err != nil {
		log.Fatalf("gagal generate key: %v", err)
	}

	ms, err := mssql.New(*connString)
	if err != nil {
		log.Fatalf("gagal konek database sqlserver: %v", err)
	}
	defer ms.Close()

	err = ms.CreateAPIKey(context.Background(), hash, model.APIKey{
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