# authz-engine

Embeddable authorization engine untuk Go yang mendukung **RBAC**, **ABAC**, dan **ReBAC** dalam satu titik keputusan melalui `Can()`.

`authz-engine` dapat digunakan sebagai:

* Go library
* HTTP authorization service
* Authorization backend dengan SQL Server
* In-memory engine untuk testing dan development

---

## Daftar Isi

* [Fitur](#fitur)
* [Arsitektur](#arsitektur)
* [Instalasi](#instalasi)
* [Setup Database](#setup-database)
* [Migration](#migration)
* [Menjalankan Server](#menjalankan-server)
* [Membuat API Key](#membuat-api-key)
* [Pemakaian sebagai Go Library](#pemakaian-sebagai-go-library)
* [HTTP API](#http-api)
* [Autentikasi dan Rate Limiting](#autentikasi-dan-rate-limiting)
* [Observability](#observability)
* [Testing dan Benchmark](#testing-dan-benchmark)
* [Konfigurasi Production](#konfigurasi-production)
* [Roadmap](#roadmap)

---

## Fitur

### RBAC

Role-based access control dengan permission dan wildcard.

Contoh:

```text
invoice:read
invoice:write
invoice:*
```

### ABAC

Role dapat memiliki `Conditions` yang dievaluasi berdasarkan context request.

Contoh:

```json
{
  "same_department": "true"
}
```

### ReBAC

Relationship-based access control dengan model relasi seperti Google Zanzibar.

Mendukung:

* Direct relation
* Hierarki relation
* Userset/group
* Relationship inheritance

Contoh:

```text
owner → editor → viewer
group:eng#member
```

### Unified `Can()`

`Can()` dapat mengevaluasi:

* RBAC
* ABAC
* ReBAC
* kombinasi RBAC + ABAC + ReBAC

Authorization menggunakan logika OR antar jalur. Jika salah satu jalur memberikan izin, request dianggap allowed.

### Caching

TTL cache opsional untuk:

* `Can()`
* `CheckRelation()`

Cache diinvalidasi ketika terjadi perubahan data yang relevan.

### HTTP API

HTTP API menggunakan `net/http` standar tanpa router pihak ketiga.

### API Key Authentication

Setiap client memiliki API key sendiri.

API key:

* disimpan sebagai SHA-256 hash
* dapat diaktifkan/dinonaktifkan
* memiliki rate limit masing-masing

### Audit Log

Perubahan state dicatat ke database, termasuk:

* create role
* assign role
* revoke role
* create relation
* delete relation
* set attribute

### Observability

Menyediakan:

* structured logging menggunakan `log/slog`
* decision log
* Prometheus metrics
* database connection pool statistics

### Database Migration

Schema database dikelola menggunakan `golang-migrate`.

Tidak diperlukan pembuatan atau perubahan schema secara manual melalui SSMS setelah migration digunakan.

### Storage

Backend yang tersedia:

* In-memory — testing/development
* SQL Server — persistent storage

Backend lain dapat ditambahkan dengan mengimplementasikan `store.Store`.

---

## Arsitektur

```text
authz-engine/
├── cmd/
│   ├── server/              # HTTP server
│   └── genkey/              # CLI untuk membuat API key
│
├── internal/
│   ├── model/               # Model inti
│   ├── store/               # Store interface + implementations
│   │   ├── memory/
│   │   └── mssql/
│   ├── engine/              # Logic authorization
│   ├── cache/               # Generic TTL cache
│   ├── api/                 # HTTP handlers, middleware, metrics
│   └── migrate/             # Migration wrapper
│
├── migrations/              # Database migrations
├── examples/                # Contoh penggunaan
├── go.mod
└── README.md
```

### Prinsip desain

`engine` hanya bergantung pada interface `store.Store`, bukan implementasi database tertentu.

Engine juga tidak bergantung langsung pada library metrics. Observability menggunakan callback `DecisionHook`.

Dengan desain ini:

* storage dapat diganti tanpa mengubah authorization logic
* metrics backend dapat diganti
* engine dapat digunakan tanpa HTTP API
* engine dapat digunakan dengan in-memory store untuk testing

---

## Instalasi

### Prasyarat

* Go 1.22+
* SQL Server — hanya diperlukan jika menggunakan persistent storage
* `golang-migrate` CLI — diperlukan untuk menjalankan migration dari terminal

Download Go:

```text
https://go.dev/dl/
```

### Clone Repository

```bash
git clone https://github.com/amayones/authz-engine.git
cd authz-engine
```

### Install Dependency

```bash
go mod tidy
```

### Build

```bash
go build ./...
```

---

# Setup Database

## 1. Buat Database

Buat database kosong di SQL Server.

Contoh:

```sql
CREATE DATABASE authzdb;
```

Pastikan user yang digunakan aplikasi memiliki permission terhadap database tersebut.

---

## 2. Connection String

Format connection string:

```text
sqlserver://USER:PASSWORD@HOST:1433?database=DATABASE&encrypt=true&trustservercertificate=true
```

Contoh:

```text
sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
```

### Penting

Jangan melakukan escaping pada `:` atau `@`.

Benar:

```text
sqlserver://may:may@localhost:1433?database=authzdb
```

Salah:

```text
sqlserver://may\:may\@localhost:1433?database=authzdb
```

Connection string juga tidak boleh memiliki karakter newline di bagian akhir.

---

# Migration

Schema database dikelola menggunakan `golang-migrate`.

Migration berada di:

```text
migrations/
```

Contoh:

```text
migrations/
├── 000001_init.up.sql
├── 000001_init.down.sql
├── 000002_add_x.up.sql
└── 000002_add_x.down.sql
```

## Install `migrate`

Install CLI:

```bash
go install -tags 'sqlserver' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Pastikan command tersedia:

```bash
migrate -version
```

---

## Set Connection String

### Git Bash

Gunakan single quote agar connection string diteruskan apa adanya:

```bash
export AUTHZ_DB_CONN='sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true'
```

Cek:

```bash
echo "$AUTHZ_DB_CONN"
```

Output harus berupa:

```text
sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
```

---

## Menjalankan Semua Migration

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations up
```

Perintah ini menjalankan seluruh migration yang belum diterapkan.

---

## Melihat Versi Migration

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations version
```

Contoh:

```text
3
```

Jika database masih kosong, version table akan dibuat otomatis oleh migrate.

---

## Rollback Satu Migration

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations down 1
```

Contoh:

```text
3 → 2
```

---

## Rollback Semua Migration

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations down
```

Gunakan dengan hati-hati karena seluruh schema yang dibuat migration akan dihapus.

---

## Membuat Migration Baru

Jika ingin menambahkan perubahan schema:

```bash
migrate create -ext sql -dir migrations -seq nama_perubahan
```

Contoh:

```bash
migrate create -ext sql -dir migrations -seq add_client_description
```

Akan menghasilkan:

```text
migrations/
├── 000001_init.up.sql
├── 000001_init.down.sql
├── 000002_add_client_description.up.sql
└── 000002_add_client_description.down.sql
```

Isi `.up.sql` dengan perubahan schema.

Contoh:

```sql
ALTER TABLE api_keys
ADD description NVARCHAR(255) NULL;
```

Isi `.down.sql` dengan rollback:

```sql
ALTER TABLE api_keys
DROP COLUMN description;
```

Kemudian jalankan:

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations up
```

---

## Force Migration Version

Gunakan `force` **hanya jika migration state di database tidak sesuai dengan file migration**.

Contoh:

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations force 1
```

`force` hanya mengubah version migration dan tidak menjalankan SQL migration.

Ini berguna jika:

* database pernah dibuat manual
* migration pernah gagal
* migration version perlu disinkronkan
* schema database sudah sesuai tetapi migration table belum sesuai

Setelah menggunakan `force`, pastikan schema database benar sebelum menjalankan migration berikutnya.

---

# Menjalankan Server

Set environment variable:

```bash
export AUTHZ_DB_CONN='sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true'
export AUTHZ_ADDR=':8080'
export AUTHZ_AUTO_MIGRATE='false'
```

Kemudian:

```bash
go run ./cmd/server
```

Server menggunakan:

```text
AUTHZ_ADDR
```

dengan default:

```text
:8080
```

### Auto Migration

Untuk menjalankan migration otomatis saat server startup:

```bash
export AUTHZ_AUTO_MIGRATE='true'
go run ./cmd/server
```

Untuk production, migration sebaiknya dijalankan sebagai bagian dari proses deployment, bukan saat setiap instance server startup.

---

## Health Check

Endpoint health check tidak membutuhkan API key:

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "status": "ok"
}
```

---

# Membuat API Key

API key pertama dibuat melalui CLI:

```bash
go run ./cmd/genkey \
  -db "$AUTHZ_DB_CONN" \
  -name "amayones" \
  -rpm 120
```

Atau satu baris:

```bash
go run ./cmd/genkey -db "$AUTHZ_DB_CONN" -name "amayones" -rpm 120
```

API key mentah hanya ditampilkan sekali.

Simpan key tersebut dengan aman.

Server hanya menyimpan hash SHA-256 dari API key.

---

# Pemakaian sebagai Go Library

Contoh penggunaan:

```go
package main

import (
    "context"
    "fmt"

    "github.com/amayones/authz-engine/internal/engine"
    "github.com/amayones/authz-engine/internal/model"
    "github.com/amayones/authz-engine/internal/store/memory"
)

func main() {
    ctx := context.Background()

    e := engine.New(memory.New())

    e.SetSchema(model.RelationSchema{
        "viewer": {"editor", "owner"},
        "editor": {"owner"},
    })

    _ = e.CreateRoleWithConditions(ctx, model.Role{
        Name: "editor",
        Permissions: []model.Permission{
            "invoice:read",
            "invoice:write",
        },
    })

    _ = e.AssignRole(ctx, "user:alice", "editor")

    allowed, _ := e.Can(ctx, model.AccessRequest{
        SubjectID: "user:alice",
        Resource:  "invoice",
        Action:    "read",
    })

    fmt.Println("allowed:", allowed)
}
```

Untuk contoh tambahan, lihat:

```text
examples/
```

---

# HTTP API

Semua endpoint membutuhkan header:

```http
X-API-Key: ak_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

kecuali:

```text
GET /health
GET /metrics
```

Body dan response menggunakan JSON.

---

## POST `/roles`

Membuat role.

Request:

```json
{
  "name": "editor",
  "permissions": [
    "invoice:read",
    "invoice:write"
  ],
  "conditions": [
    {
      "attr_key": "same_department",
      "attr_value": "true"
    }
  ]
}
```

Response:

```json
{
  "status": "created"
}
```

Status:

```text
201 Created
```

---

## POST `/roles/assign`

Assign role ke subject.

```json
{
  "subject_id": "user:alice",
  "role_name": "editor"
}
```

---

## POST `/roles/revoke`

Mencabut role dari subject.

```json
{
  "subject_id": "user:alice",
  "role_name": "editor"
}
```

---

## POST `/can`

Endpoint utama authorization.

### RBAC / ABAC

```json
{
  "subject_id": "user:alice",
  "resource": "invoice",
  "action": "read",
  "context": {
    "same_department": "true"
  }
}
```

### ReBAC

Isi `object` untuk melakukan pengecekan relation.

```json
{
  "subject_id": "user:bob",
  "action": "viewer",
  "object": "document:123"
}
```

Response:

```json
{
  "allowed": true
}
```

---

## POST `/relations`

Membuat relationship tuple.

```json
{
  "object": "document:123",
  "relation": "owner",
  "subject": "user:alice"
}
```

---

## DELETE `/relations`

Menghapus relationship tuple.

```json
{
  "object": "document:123",
  "relation": "owner",
  "subject": "user:alice"
}
```

---

## POST `/relations/check`

Mengecek ReBAC secara langsung tanpa RBAC.

```json
{
  "object": "document:123",
  "relation": "viewer",
  "subject": "user:bob"
}
```

Response:

```json
{
  "allowed": true
}
```

---

## POST `/attributes`

Menyimpan attribute subject untuk kebutuhan ABAC.

```json
{
  "subject_id": "user:alice",
  "key": "department",
  "value": "engineering"
}
```

---

## GET `/health`

Health check tanpa API key.

---

## GET `/metrics`

Prometheus metrics tanpa API key.

Jika server terekspos ke jaringan publik, endpoint ini sebaiknya dibatasi menggunakan firewall atau network policy.

---

## Error Response

Format error:

```json
{
  "error": "pesan error"
}
```

Status code:

| Status | Keterangan                         |
| ------ | ---------------------------------- |
| `400`  | Request tidak valid                |
| `401`  | API key tidak valid atau tidak ada |
| `409`  | Data sudah ada                     |
| `429`  | Rate limit terlampaui              |
| `500`  | Internal server error              |

---

# Autentikasi dan Rate Limiting

Setiap client memiliki API key sendiri.

API key:

* dibuat melalui `cmd/genkey`
* disimpan sebagai SHA-256 hash
* tidak disimpan dalam bentuk plaintext
* dapat dinonaktifkan tanpa menghapus history

Rate limit dikonfigurasi saat membuat API key:

```bash
go run ./cmd/genkey \
  -db "$AUTHZ_DB_CONN" \
  -name "amayones" \
  -rpm 120
```

Gunakan API key melalui:

```http
X-API-Key: ak_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Rate limiting menggunakan token bucket dan diterapkan per API key.

---

# Observability

## Audit Log

Perubahan state dicatat ke tabel:

```text
audit_log
```

Contoh query:

```sql
SELECT *
FROM audit_log
ORDER BY occurred_at DESC;
```

Audit log mencatat:

* actor
* action
* target
* timestamp

---

## Decision Log

Setiap evaluasi:

```text
Can()
CheckRelation()
```

dicatat sebagai structured log menggunakan `log/slog`.

Log mencakup informasi seperti:

* subject
* action
* object/resource
* hasil authorization
* cache hit/miss

Log dapat dikirim ke sistem seperti Loki atau ELK.

---

## Prometheus Metrics

Endpoint:

```text
GET /metrics
```

Metrics utama:

| Metric                                | Keterangan                                                  |
| ------------------------------------- | ----------------------------------------------------------- |
| `authz_http_requests_total`           | Total HTTP request berdasarkan endpoint, method, dan status |
| `authz_http_request_duration_seconds` | Histogram latency HTTP request                              |
| `authz_decisions_total`               | Total authorization decision berdasarkan jenis dan hasil    |
| `authz_cache_hits_total`              | Cache hit dan miss                                          |
| `authz_rate_limit_rejected_total`     | Request yang ditolak karena rate limit                      |

---

# Testing dan Benchmark

Build:

```bash
go build ./...
```

Static analysis:

```bash
go vet ./...
```

Test:

```bash
go test ./... -v -race
```

Benchmark:

```bash
go test ./internal/engine -bench=. -benchmem -run=^$
```

Unit test engine menggunakan in-memory store sehingga tidak membutuhkan SQL Server.

Untuk integration test SQL Server, gunakan database test terpisah dan set:

```bash
export AUTHZ_DB_CONN='sqlserver://user:password@localhost:1433?database=authz_test&encrypt=true&trustservercertificate=true'
```

---

# Konfigurasi Production

| Environment Variable | Default                                               | Keterangan                         |
| -------------------- | ----------------------------------------------------- | ---------------------------------- |
| `AUTHZ_DB_CONN`      | `sqlserver://may:may@localhost:1433?database=authzdb` | SQL Server connection string       |
| `AUTHZ_ADDR`         | `:8080`                                               | HTTP listen address                |
| `AUTHZ_AUTO_MIGRATE` | `false`                                               | Menjalankan migration saat startup |

### Connection Pool

Default:

```text
MaxOpenConns:      25
MaxIdleConns:      10
ConnMaxLifetime:   30m
ConnMaxIdleTime:   5m
```

Gunakan `NewWithConfig()` jika membutuhkan tuning khusus.

### Request Timeout

HTTP request memiliki timeout 8 detik.

### TLS

`authz-engine` tidak menangani HTTPS secara langsung.

Untuk production, jalankan di belakang reverse proxy seperti:

```text
Client
  ↓
HTTPS
  ↓
Nginx / Caddy
  ↓
authz-engine :8080
```

### Migration

Untuk production, jalankan migration sebagai langkah deployment:

```bash
migrate -database "$AUTHZ_DB_CONN" -path ./migrations up
```

Kemudian jalankan server:

```bash
go run ./cmd/server
```

Hindari mengandalkan `AUTHZ_AUTO_MIGRATE=true` pada deployment production multi-instance karena setiap instance dapat mencoba melakukan migration saat startup.

### Metrics

Endpoint `/metrics` tidak membutuhkan API key.

Batasi aksesnya menggunakan:

* firewall
* private network
* reverse proxy
* network policy

---
