# authz-engine

**Authorization engine untuk Go** yang mendukung **RBAC**, **ABAC**, dan **ReBAC** dalam satu titik keputusan melalui `Can()`.

Bisa dipakai sebagai:
- **Go library** — embed langsung di aplikasi Anda
- **HTTP service** — REST API dengan autentikasi API key
- **Backend SQL Server** — penyimpanan persisten

---

## 📋 Daftar Isi

- [Fitur](#fitur)
- [Konsep Dasar](#konsep-dasar)
- [Arsitektur](#arsitektur)
- [Instalasi](#instalasi)
- [Menjalankan Server](#menjalankan-server)
- [Membuat API Key](#membuat-api-key)
- [HTTP API](#http-api)
- [Pemakaian sebagai Go Library](#pemakaian-sebagai-go-library)
- [Konfigurasi](#konfigurasi)
- [Observability](#observability)
- [Testing](#testing)
- [Deployment](#deployment)

---

## ✨ Fitur

### RBAC (Role-Based Access Control)

Kontrol akses berbasis **role**. Setiap role punya daftar permission.

```text
invoice:read      → boleh baca invoice
invoice:write     → boleh tulis invoice
invoice:*         → boleh semua aksi pada invoice (wildcard)
```

### ABAC (Attribute-Based Access Control)

Role bisa punya **kondisi** yang dievaluasi berdasarkan context request.

```json
{
  "same_department": "true"
}
```

Artinya: role hanya berlaku jika context request punya `same_department=true`.

### ReBAC (Relationship-Based Access Control)

Kontrol akses berbasis **relasi** antar objek (model Google Zanzibar).

Mendukung:
- Direct relation (langsung)
- Hierarki relation (parent/child)
- Userset/group (keanggotaan grup)
- Relationship inheritance (pewarisan relasi)

```text
owner → editor → viewer
group:eng#member
```

### Unified `Can()`

Satu fungsi `Can()` bisa mengevaluasi **RBAC + ABAC + ReBAC sekaligus**.
Jika salah satu jalur memberi izin, request dianggap **allowed**.

### Caching

TTL cache opsional untuk mempercepat `Can()` dan `CheckRelation()`.
Cache otomatis diinvalidasi saat data berubah.

### API Key Authentication

Setiap client punya API key sendiri:
- Disimpan sebagai **SHA-256 hash** (bukan plaintext)
- Bisa diaktifkan/dinonaktifkan
- Punya **rate limit** masing-masing

### Audit Log

Semua perubahan state dicatat ke database:
- create role
- assign role
- revoke role
- create relation
- delete relation
- set attribute

### Observability

- Structured logging (`log/slog`)
- Decision log
- Prometheus metrics
- Database connection pool stats

---

## 🧠 Konsep Dasar

### Subject

Entitas yang meminta akses — biasanya user, tapi bisa juga service-account.

```text
user:alice
user:bob
service:billing
```

### Permission

Satu izin atomik dalam format `resource:action`.

```text
invoice:read
invoice:write
invoice:delete
```

### Role

Kumpulan permission yang bisa di-assign ke subject.

```go
model.Role{
    Name: "editor",
    Permissions: []model.Permission{
        "invoice:read",
        "invoice:write",
    },
}
```

### AccessRequest

Pertanyaan: *"Apakah subject ini boleh melakukan action pada resource ini?"*

```go
model.AccessRequest{
    SubjectID: "user:alice",
    Resource:  "invoice",
    Action:    "read",
}
```

### RelationTuple

Unit dasar ReBAC: pernyataan bahwa `Subject` punya `Relation` terhadap `Object`.

```text
document:123  ←  owner  ←  user:alice
```

### RelationSchema

Hierarki relasi. Jika `viewer` berisi `editor` dan `owner`, maka siapa pun yang punya relasi `editor` atau `owner` otomatis dianggap `viewer`.

```go
model.RelationSchema{
    "viewer": {"editor", "owner"},
    "editor": {"owner"},
}
```

---

## 🏗️ Arsitektur

```text
authz-engine/
├── cmd/
│   ├── server/              # HTTP server (main entry)
│   └── genkey/              # CLI untuk membuat API key
│
├── internal/
│   ├── model/               # Tipe data inti (Role, Permission, dll.)
│   ├── store/               # Interface penyimpanan + implementasi
│   │   ├── memory/          # In-memory store (testing/development)
│   │   └── mssql/           # SQL Server store (production)
│   ├── engine/              # Logic authorization (Can, CheckRelation)
│   ├── cache/               # Generic TTL cache
│   ├── api/                 # HTTP handlers, middleware, metrics
│   └── migrate/             # Database migration wrapper
│
├── migrations/              # SQL migration files
├── examples/                # Contoh penggunaan
├── start.bat                # Script start server (Windows)
├── stop.bat                 # Script stop server (Windows)
├── tunnel.bat               # Script tunnel publik (Windows)
├── DEPLOY.md                # Panduan deployment & self-hosting
└── go.mod
```

### Prinsip Desain

- `engine` hanya bergantung pada **interface** `store.Store`, bukan implementasi database tertentu
- Storage bisa diganti tanpa mengubah logic authorization
- Engine bisa dipakai tanpa HTTP API
- Engine bisa dipakai dengan in-memory store untuk testing

---

## 🚀 Instalasi

### Prasyarat

- **Go 1.22+** — https://go.dev/dl/
- **SQL Server** — hanya jika pakai persistent storage

### Clone & Build

```bash
git clone https://github.com/amayones/authz-engine.git
cd authz-engine

go mod tidy
go build ./...
```

---

## ▶️ Menjalankan Server

### Cara 1: Pakai `start.bat` (Windows, paling mudah)

```bash
start.bat
```

Script ini otomatis:
1. Build `authz-server.exe` jika belum ada
2. Set environment variables
3. Jalankan server di `http://localhost:8080`

### Cara 2: Manual

```bash
# Windows (Command Prompt)
set AUTHZ_DB_DRIVER=sqlserver
set AUTHZ_DB_CONN=sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
set AUTHZ_ADDR=:8080
set AUTHZ_AUTO_MIGRATE=true
go run ./cmd/server
```

```bash
# Linux / Mac
export AUTHZ_DB_DRIVER=sqlserver
export AUTHZ_DB_CONN='sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true'
export AUTHZ_ADDR=':8080'
export AUTHZ_AUTO_MIGRATE='true'
go run ./cmd/server
```

### Verifikasi

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"ok"}
```

---

## 🔑 Membuat API Key

API key dibuat via CLI `cmd/genkey`:

```bash
go run ./cmd/genkey \
  -db "sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true" \
  -name "client1" \
  -rpm 120
```

Output:
```
API key berhasil dibuat. SIMPAN SEKARANG — tidak akan ditampilkan lagi:
ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```

> **PENTING**: API key hanya ditampilkan **sekali**. Server hanya menyimpan hash SHA-256.

---

## 🌐 HTTP API

Semua endpoint membutuhkan header:

```http
X-API-Key: ak_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Kecuali:
- `GET /health` — health check
- `GET /metrics` — Prometheus metrics

### POST `/roles` — Buat role

```json
{
  "name": "editor",
  "permissions": ["invoice:read", "invoice:write"],
  "conditions": [
    {"attr_key": "same_department", "attr_value": "true"}
  ]
}
```

Response: `201 Created`

### POST `/roles/assign` — Assign role ke user

```json
{
  "subject_id": "user:alice",
  "role_name": "editor"
}
```

### POST `/roles/revoke` — Cabut role dari user

```json
{
  "subject_id": "user:alice",
  "role_name": "editor"
}
```

### POST `/can` — Cek izin (endpoint utama)

**RBAC / ABAC:**
```json
{
  "subject_id": "user:alice",
  "resource": "invoice",
  "action": "read",
  "context": {"same_department": "true"}
}
```

**ReBAC:**
```json
{
  "subject_id": "user:bob",
  "action": "viewer",
  "object": "document:123"
}
```

Response:
```json
{"allowed": true}
```

### POST `/relations` — Buat relasi ReBAC

```json
{
  "object": "document:123",
  "relation": "owner",
  "subject": "user:alice"
}
```

### DELETE `/relations` — Hapus relasi ReBAC

```json
{
  "object": "document:123",
  "relation": "owner",
  "subject": "user:alice"
}
```

### POST `/relations/check` — Cek relasi ReBAC langsung

```json
{
  "object": "document:123",
  "relation": "viewer",
  "subject": "user:bob"
}
```

Response:
```json
{"allowed": true}
```

### POST `/attributes` — Set atribut subject (untuk ABAC)

```json
{
  "subject_id": "user:alice",
  "key": "department",
  "value": "engineering"
}
```

### Error Response

```json
{"error": "pesan error"}
```

| Status | Keterangan |
|--------|------------|
| `400` | Request tidak valid |
| `401` | API key tidak valid |
| `409` | Data sudah ada |
| `429` | Rate limit terlampaui |
| `500` | Internal server error |

---

## 📦 Pemakaian sebagai Go Library

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

    // Buat engine dengan in-memory store
    e := engine.New(memory.New())

    // Set hierarki relasi ReBAC
    e.SetSchema(model.RelationSchema{
        "viewer": {"editor", "owner"},
        "editor": {"owner"},
    })

    // Buat role
    _ = e.CreateRoleWithConditions(ctx, model.Role{
        Name: "editor",
        Permissions: []model.Permission{
            "invoice:read",
            "invoice:write",
        },
    })

    // Assign role ke user
    _ = e.AssignRole(ctx, "user:alice", "editor")

    // Cek izin
    allowed, _ := e.Can(ctx, model.AccessRequest{
        SubjectID: "user:alice",
        Resource:  "invoice",
        Action:    "read",
    })

    fmt.Println("allowed:", allowed) // allowed: true
}
```

Contoh lain ada di folder `examples/`:
- `examples/basic/` — RBAC sederhana
- `examples/rebac/` — ReBAC dengan relasi
- `examples/unified/` — Kombinasi RBAC + ABAC + ReBAC

---

## ⚙️ Konfigurasi

### Environment Variables

| Variable | Default | Keterangan |
|----------|---------|------------|
| `AUTHZ_DB_DRIVER` | `sqlserver` | Driver database |
| `AUTHZ_DB_CONN` | `sqlserver://may:may@localhost:1433?database=authzdb` | Connection string SQL Server |
| `AUTHZ_ADDR` | `:8080` | HTTP listen address |
| `AUTHZ_AUTO_MIGRATE` | `false` | Jalankan migration saat startup |

### Connection String SQL Server

```text
sqlserver://USER:PASSWORD@HOST:1433?database=DATABASE&encrypt=true&trustservercertificate=true
```

Contoh:
```text
sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
```

> **PENTING**: Jangan escape `:` atau `@` di connection string.

### Connection Pool

| Setting | Default |
|---------|---------|
| MaxOpenConns | 25 |
| MaxIdleConns | 10 |
| ConnMaxLifetime | 30m |
| ConnMaxIdleTime | 5m |

### Request Timeout

HTTP request memiliki timeout 8 detik.

---

## 📊 Observability

### Audit Log

Semua perubahan state dicatat ke tabel `audit_log`:

```sql
SELECT * FROM audit_log ORDER BY occurred_at DESC;
```

### Decision Log

Setiap evaluasi `Can()` dan `CheckRelation()` dicatat sebagai structured log (`log/slog`).

### Prometheus Metrics

Endpoint: `GET /metrics`

| Metric | Keterangan |
|--------|------------|
| `authz_http_requests_total` | Total HTTP request |
| `authz_http_request_duration_seconds` | Histogram latency |
| `authz_decisions_total` | Total keputusan authorization |
| `authz_cache_hits_total` | Cache hit/miss |
| `authz_rate_limit_rejected_total` | Request ditolak rate limit |

---

## 🧪 Testing

```bash
# Build
go build ./...

# Static analysis
go vet ./...

# Test (dengan race detector)
go test ./... -v -race

# Benchmark
go test ./internal/engine -bench=. -benchmem -run=^$
```

Unit test engine menggunakan in-memory store — **tidak butuh SQL Server**.

---

## 🚢 Deployment

### Self-Hosting (PC/Laptop)

Panduan lengkap ada di **`DEPLOY.md`**:

1. **Build**: `go build -o authz-server.exe ./cmd/server`
2. **Jalankan**: `start.bat`
3. **Akses publik**: `tunnel.bat` (Cloudflare/ngrok)

### Script yang Tersedia

| Script | Fungsi |
|--------|--------|
| `start.bat` | Jalankan server (build otomatis + set env) |
| `stop.bat` | Stop server |
| `tunnel.bat` | Tunnel publik (Cloudflare/ngrok) |

### Docker

```bash
docker build -t authz-engine .
docker run -p 8080:8080 \
  -e AUTHZ_DB_CONN="sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true" \
  authz-engine
```

---

## 📄 Lisensi

Proyek ini open-source. Silakan gunakan, modifikasi, dan distribusikan sesuai kebutuhan.