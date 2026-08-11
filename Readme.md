# authz-engine

Embeddable authorization engine untuk Go — mendukung **RBAC**, **ABAC**, dan **ReBAC**
(model relasi ala Google Zanzibar) dalam satu titik keputusan (`Can()`). Bisa dipakai
langsung sebagai Go library, atau lewat HTTP API supaya bisa dipanggil dari bahasa apa saja.

## Daftar Isi

- [Fitur](#fitur)
- [Arsitektur](#arsitektur)
- [Instalasi](#instalasi)
- [Setup Database](#setup-database)
- [Menjalankan Server](#menjalankan-server)
- [Pemakaian sebagai Go Library](#pemakaian-sebagai-go-library)
- [Referensi HTTP API](#referensi-http-api)
- [Autentikasi & Rate Limiting](#autentikasi--rate-limiting)
- [Observability](#observability)
- [Migration](#migration)
- [Testing & Benchmark](#testing--benchmark)
- [Konfigurasi Production](#konfigurasi-production)

---

## Fitur

- **RBAC** — role dengan permission, mendukung wildcard (`invoice:*`)
- **ABAC** — role bisa punya `Conditions` yang dievaluasi terhadap context request
- **ReBAC** — relationship graph mirip Google Zanzibar: direct match, hierarki relasi
  (`owner` → `editor` → `viewer`), dan userset/grup (`group:eng#member`)
- **`Can()` terpadu** — satu fungsi yang mengevaluasi RBAC+ABAC dan (opsional) ReBAC
  sekaligus, logika OR: cukup salah satu jalur mengizinkan
- **Caching** — TTL cache opsional untuk hasil `Can()`/`CheckRelation()`, dengan
  invalidasi otomatis saat ada mutasi data
- **HTTP API** — 9 endpoint di atas `net/http` standar (tanpa router pihak ketiga)
- **Auth per-client** — API key per client (disimpan sebagai hash), rate limit per key
- **Audit log** — semua perubahan state (assign role, write relation, dll) tercatat
- **Decision log & metrics** — structured logging (`slog`) + Prometheus metrics
- **Schema migration** — lewat `golang-migrate`, bukan lagi manual via GUI client

Storage backend: in-memory (untuk testing/development) atau SQL Server (untuk
persistence). Backend lain tinggal implementasi interface `store.Store`.

## Arsitektur

```
authz-engine/
├── cmd/
│   ├── server/        entry point HTTP server
│   └── genkey/         CLI admin untuk provisioning API key
├── internal/
│   ├── model/           tipe data inti (Role, Permission, RelationTuple, dst)
│   ├── store/            interface abstrak + implementasi memory & mssql
│   ├── engine/           logic inti: Can(), CheckRelation(), caching, audit hook
│   ├── cache/            TTL cache generik
│   ├── api/              HTTP handlers, middleware, metrics
│   └── migrate/          wrapper golang-migrate
├── migrations/           file .up.sql / .down.sql
└── examples/             contoh pemakaian sebagai Go library
```

Prinsip desain: `engine` tidak pernah depend ke `store` konkret (cuma interface), dan
tidak depend ke library metrics tertentu (pakai `DecisionHook` callback). Jadi backend
storage atau sistem metrics bisa diganti tanpa menyentuh logic inti.

## Instalasi

### Prasyarat

- Go 1.22 atau lebih baru — [https://go.dev/dl](https://go.dev/dl)
- Akses ke SQL Server (opsional untuk development; in-memory store cukup untuk
  eksplorasi awal) — bisa server existing kantor/cloud, atau install lokal
- [`golang-migrate` CLI](https://github.com/golang-migrate/migrate) (opsional, hanya
  kalau mau jalankan migration manual dari terminal)

### Clone & Build

```bash
git clone https://github.com/amayones/authz-engine.git
cd authz-engine
go mod tidy
go build ./...
```

## Setup Database

1. Siapkan database kosong di SQL Server, misal `authzdb`.
2. Jalankan migration (lihat [Migration](#migration)) untuk membuat seluruh skema:
   `roles`, `subject_roles`, `subject_attributes`, `relation_tuples`, `api_keys`,
   `audit_log`.
3. Siapkan connection string:

   ```
   sqlserver://<user>:<password>@<host>:1433?database=authzdb&encrypt=true&trustservercertificate=true
   ```

   Parameter `trustservercertificate=true` diperlukan kalau server memakai
   self-signed certificate (umum untuk server internal). Hilangkan parameter ini
   kalau server sudah punya certificate yang valid.

## Menjalankan Server

```bash
export AUTHZ_DB_CONN="sqlserver://may:may@HOST_SERVER:1433?database=authzdb&encrypt=true&trustservercertificate=true"
export AUTHZ_ADDR=":8080"
export AUTHZ_AUTO_MIGRATE="false"   # set "true" untuk auto-migrate saat startup

go run ./cmd/server
```

Server akan listen di `AUTHZ_ADDR` (default `:8080`). Endpoint `/health` bisa dicek
tanpa API key:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Membuat API Key Pertama

API key **tidak** dibuat lewat HTTP (supaya tidak sembarang orang bisa provisioning
key sendiri), tapi lewat CLI admin:

```bash
go run ./cmd/genkey \
  -db "$AUTHZ_DB_CONN" \
  -name "nama-client-kamu" \
  -rpm 120
```

Output key mentah **hanya ditampilkan sekali** — simpan segera, server hanya
menyimpan hash-nya.

## Pemakaian sebagai Go Library

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
	e := engine.New(memory.New()) // atau engine.NewWithCache(store, ttl) untuk caching

	e.SetSchema(model.RelationSchema{
		"viewer": {"editor", "owner"},
		"editor": {"owner"},
	})

	_ = e.CreateRoleWithConditions(ctx, model.Role{
		Name:        "editor",
		Permissions: []model.Permission{"invoice:read", "invoice:write"},
	})
	_ = e.AssignRole(ctx, "user:alice", "editor")

	allowed, _ := e.Can(ctx, model.AccessRequest{
		SubjectID: "user:alice",
		Resource:  "invoice",
		Action:    "read",
	})
	fmt.Println("allowed:", allowed) // true
}
```

Lihat folder `examples/` untuk skenario lain: RBAC murni, ReBAC dengan userset/grup,
dan kombinasi RBAC+ReBAC dalam satu `Can()`.

## Referensi HTTP API

Semua endpoint (kecuali `/health` dan `/metrics`) butuh header `X-API-Key`.
Body & response berformat JSON.

### `POST /roles` — Buat role baru

```json
// Request
{
  "name": "editor",
  "permissions": ["invoice:read", "invoice:write"],
  "conditions": [{"attr_key": "same_department", "attr_value": "true"}]
}
```
```json
// Response 201
{"status": "created"}
```

### `POST /roles/assign` — Assign role ke subject

```json
{"subject_id": "user:alice", "role_name": "editor"}
```

### `POST /roles/revoke` — Cabut role dari subject

```json
{"subject_id": "user:alice", "role_name": "editor"}
```

### `POST /can` — Titik keputusan utama (RBAC + ABAC + ReBAC)

```json
// Request — RBAC/ABAC saja (Object dikosongkan)
{
  "subject_id": "user:alice",
  "resource": "invoice",
  "action": "read",
  "context": {"same_department": "true"}
}
```
```json
// Request — ikut cek ReBAC (Object diisi)
{
  "subject_id": "user:bob",
  "action": "viewer",
  "object": "document:123"
}
```
```json
// Response
{"allowed": true}
```

### `POST /relations` — Buat relation tuple (ReBAC)

```json
{"object": "document:123", "relation": "owner", "subject": "user:alice"}
```

### `DELETE /relations` — Hapus relation tuple

```json
{"object": "document:123", "relation": "owner", "subject": "user:alice"}
```

### `POST /relations/check` — Cek relasi ReBAC murni (tanpa RBAC)

```json
{"object": "document:123", "relation": "viewer", "subject": "user:bob"}
```
```json
{"allowed": true}
```

### `POST /attributes` — Set atribut subject (ABAC)

```json
{"subject_id": "user:alice", "key": "department", "value": "engineering"}
```

### `GET /health` — Health check (tanpa API key)

### `GET /metrics` — Prometheus metrics (tanpa API key, sebaiknya dibatasi firewall)

### Format Error

```json
{"error": "pesan error di sini"}
```

Status code: `400` (input tidak valid), `401` (API key salah/tidak ada), `409` (data
sudah ada), `429` (rate limit), `500` (error internal).

## Autentikasi & Rate Limiting

- Tiap client punya API key sendiri, di-generate lewat `cmd/genkey`, disimpan sebagai
  **hash SHA-256** di tabel `api_keys` — server tidak pernah menyimpan key mentah.
- Rate limit **per client** (token bucket), diatur lewat kolom `rate_limit_rpm` saat
  key dibuat (`-rpm` flag di `genkey`).
- Key bisa dicabut (`is_active = 0`) tanpa menghapus riwayatnya.
- Kirim key lewat header:

  ```
  X-API-Key: ak_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  ```

## Observability

### Audit Log

Setiap perubahan state (create role, assign/revoke, write/delete relation, set
attribute) tercatat di tabel `audit_log`: siapa (`actor`, dari nama client API key),
apa (`action`), objek yang berubah (`target`), kapan (`occurred_at`).

```sql
SELECT * FROM audit_log ORDER BY occurred_at DESC;
```

### Decision Log

Setiap evaluasi `Can()` / `CheckRelation()` dicatat sebagai structured JSON log ke
stdout (lewat `log/slog`) — termasuk siapa, aksi apa, hasil allow/deny, dan apakah
kena cache. Cocok dialirkan ke log aggregator (Loki, ELK, dll).

### Metrics (Prometheus)

Endpoint `GET /metrics` mengekspos:

| Metric | Deskripsi |
|---|---|
| `authz_http_requests_total` | Total request per endpoint, method, status code |
| `authz_http_request_duration_seconds` | Histogram latency per endpoint |
| `authz_decisions_total` | Total keputusan, per kind (`can`/`check_relation`) & hasil |
| `authz_cache_hits_total` | Cache hit vs miss |
| `authz_rate_limit_rejected_total` | Total request ditolak karena rate limit |

## Migration

Semua perubahan skema tercatat sebagai file di `migrations/`.

```bash
# Jalankan semua migration yang belum diterapkan
migrate -database "$AUTHZ_DB_CONN" -path ./migrations up

# Rollback satu langkah
migrate -database "$AUTHZ_DB_CONN" -path ./migrations down 1

# Buat migration baru
migrate create -ext sql -dir migrations -seq nama_perubahan
```

Untuk database yang skemanya sudah dibuat manual sebelum migration dipakai, gunakan
`force <versi>` untuk menyinkronkan tanpa re-run SQL yang sama — lihat komentar di
`migrations/000001_*` untuk konteks.

## Testing & Benchmark

```bash
go build ./...
go vet ./...
go test ./... -v -race
go test ./internal/engine -bench=. -benchmem -run=^$
```

Semua test engine memakai in-memory store (tidak butuh koneksi database). Untuk
integration test terhadap SQL Server sungguhan, siapkan database test terpisah dan
set `AUTHZ_DB_CONN` sebelum menjalankan test bertag `integration` (kalau ada,
sesuaikan dengan setup CI kamu).

## Konfigurasi Production

| Env Var | Default | Keterangan |
|---|---|---|
| `AUTHZ_DB_CONN` | `sqlserver://may:may@localhost:1433?database=authzdb` | Connection string SQL Server |
| `AUTHZ_ADDR` | `:8080` | Alamat listen HTTP server |
| `AUTHZ_AUTO_MIGRATE` | `false` | Jalankan migration otomatis saat startup |

Catatan production:

- **Connection pool** sudah diatur eksplisit (`MaxOpenConns: 25`, dst — lihat
  `internal/store/mssql/mssql_store.go`), sesuaikan lewat `NewWithConfig` kalau
  traffic sudah besar.
- **Request timeout** 8 detik per request (lihat `timeoutMiddleware`).
- **TLS**: server ini tidak menangani HTTPS langsung — taruh di belakang reverse
  proxy (nginx/Caddy) yang menangani TLS termination.
- **Auto-migrate saat startup** nyaman untuk staging, tapi untuk production
  disarankan jalankan migration sebagai langkah deploy terpisah, supaya kegagalan
  migration diketahui sebelum traffic mulai masuk ke versi baru.
- **`/metrics`** tidak diproteksi API key (asumsi discrape oleh Prometheus internal)
  — batasi lewat firewall/network policy kalau server terekspos ke internet luas.

---

## Roadmap Selanjutnya (Belum Diimplementasi)

- Integration test otomatis terhadap SQL Server sungguhan (bukan cuma memory store)
- Per-client API key rotation
- Distributed rate limiting (Redis) untuk multi-instance deployment
- gRPC API sebagai alternatif HTTP untuk kebutuhan latency lebih rendah