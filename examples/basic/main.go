package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amayones/authz-engine/internal/engine"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store/mssql"
)

func main() {
	ctx := context.Background()

	// Ganti HOST_SERVER dengan alamat server SQL Server kamu.
	connString := "sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true"
	s, err := mssql.New(connString)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	defer s.Close()

	e := engine.New(s)

	// CreateRole akan error kalau role sudah ada dari run sebelumnya —
	// aman diabaikan untuk demo.
	_ = e.CreateRole(ctx, "editor", []model.Permission{"invoice:read", "invoice:write"})
	_ = e.AssignRole(ctx, "user-may", "editor")

	allowed, err := e.Can(ctx, model.AccessRequest{
		SubjectID: "user-may",
		Resource:  "invoice",
		Action:    "read",
	})
	if err != nil {
		log.Fatalf("evaluasi gagal: %v", err)
	}
	fmt.Printf("allowed=%v\n", allowed)
}
