# Panduan Self-Hosting — Jalankan dari PC/Laptop

Panduan ini menjelaskan cara menjalankan `authz-engine` **langsung dari PC/laptop Anda** (self-host), tanpa perlu layanan cloud berbayar.

---

## Ringkasan

| Komponen | Cara |
|----------|------|
| Aplikasi Go | Jalankan langsung dari PC/laptop (Windows/Linux/Mac) |
| Database | PostgreSQL lokal (gratis) atau Supabase (gratis 500MB) |
| Akses dari luar | Port forwarding / tunnel (ngrok, cloudflared) |

---

## Opsi Database

### Opsi A: PostgreSQL Lokal (paling sederhana)

Install PostgreSQL di PC Anda:

1. Download: https://www.postgresql.org/download/
2. Install dengan default settings
3. Buat database:
   ```sql
   CREATE DATABASE authzdb;
   ```

Connection string:
```
postgresql://postgres:YOUR_PASSWORD@localhost:5432/authzdb
```

### Opsi B: Supabase (gratis, tanpa install)

1. Buka https://supabase.com → **Start your project**
2. **New project**:
   - **Name**: `authz-engine`
   - **Password**: buat password kuat
   - **Region**: **Southeast Asia (Singapore)**
3. **Settings → Database → Connection string** → pilih **URI**
4. Salin connection string:
   ```
   postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

---

## Langkah 1: Build Aplikasi

Buka terminal di folder proyek:

```bash
cd c:\Apache2462\htdocs\authz-engine
go build -o authz-server.exe ./cmd/server
go build -o authz-genkey.exe ./cmd/genkey
```

> Di Linux/Mac: `go build -o authz-server ./cmd/server`

---

## Langkah 2: Jalankan Server

### Windows (Command Prompt / PowerShell)

```cmd
set AUTHZ_DB_DRIVER=postgres
set AUTHZ_DB_CONN=postgresql://postgres:PASSWORD@localhost:5432/authzdb
set AUTHZ_ADDR=:8080
set AUTHZ_AUTO_MIGRATE=true
authz-server.exe
```

### Windows (Git Bash)

```bash
export AUTHZ_DB_DRIVER=postgres
export AUTHZ_DB_CONN='postgresql://postgres:PASSWORD@localhost:5432/authzdb'
export AUTHZ_ADDR=':8080'
export AUTHZ_AUTO_MIGRATE='true'
./authz-server.exe
```

### Linux / Mac

```bash
export AUTHZ_DB_DRIVER=postgres
export AUTHZ_DB_CONN='postgresql://postgres:PASSWORD@localhost:5432/authzdb'
export AUTHZ_ADDR=':8080'
export AUTHZ_AUTO_MIGRATE='true'
./authz-server
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

## Langkah 3: Buat API Key

```bash
go run ./cmd/genkey \
  -driver postgres \
  -db "postgresql://postgres:PASSWORD@localhost:5432/authzdb" \
  -name "client1" \
  -rpm 120
```

Output:
```
API key berhasil dibuat. SIMPAN SEKARANG — tidak akan ditampilkan lagi:
ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```

---

## Langkah 4: Uji API

### Buat role
```bash
curl -X POST http://localhost:8080/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"name": "editor", "permissions": ["invoice:read", "invoice:write"]}'
```

### Assign role
```bash
curl -X POST http://localhost:8080/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### Cek izin
```bash
curl -X POST http://localhost:8080/can \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "resource": "invoice", "action": "read"}'
```

Response:
```json
{"allowed":true}
```

---

## Akses dari Luar (Opsional)

Jika ingin diakses dari internet (mis. dari HP atau komputer lain), gunakan tunnel gratis:

### Opsi 1: Cloudflare Tunnel (gratis, tanpa install)

1. Download: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
2. Jalankan:
   ```bash
   cloudflared tunnel --url http://localhost:8080
   ```
3. Cloudflare memberi URL sementara:
   ```
   https://random-name.trycloudflare.com
   ```

### Opsi 2: ngrok (gratis)

1. Download: https://ngrok.com/download
2. Daftar & dapatkan token
3. Jalankan:
   ```bash
   ngrok http 8080
   ```
4. ngrok memberi URL:
   ```
   https://random-name.ngrok-free.app
   ```

---

## Script Otomatis (Windows)

### `start.bat` — Jalankan server

```bat
@echo off
cd /d %~dp0
set AUTHZ_DB_DRIVER=postgres
set AUTHZ_DB_CONN=postgresql://postgres:PASSWORD@localhost:5432/authzdb
set AUTHZ_ADDR=:8080
set AUTHZ_AUTO_MIGRATE=true
authz-server.exe
```

### `stop.bat` — Stop server

```bat
@echo off
taskkill /f /im authz-server.exe
```

---

## Troubleshooting

### Error: `connection refused` saat migration

- Pastikan PostgreSQL sudah berjalan
- Cek connection string benar (host, port, password)
- Jika pakai Supabase, pastikan memilih **URI** (bukan Transaction/Pooler)

### Port 8080 sudah dipakai

- Ganti port: `set AUTHZ_ADDR=:9090`
- Akses di `http://localhost:9090`

### Firewall Windows memblokir

- Saat pertama kali menjalankan, Windows akan menanyakan akses firewall
- Klik **Allow access** agar bisa diakses dari jaringan lokal

### Ingin auto-start saat Windows boot

1. Tekan `Win + R`, ketik `shell:startup`, Enter
2. Buat shortcut ke `start.bat` di folder startup

---

## Catatan Penting

- **Self-host** = aplikasi berjalan selama PC/laptop menyala
- **Database lokal** = data tersimpan di PC Anda
- **Supabase** = data tersimpan di cloud (gratis 500MB)
- **Tunnel** (ngrok/cloudflared) = akses dari internet tanpa port forwarding manual
- **Auto-start** = aplikasi jalan otomatis saat Windows boot

---

## File yang Tidak Diperlukan (Sudah Dihapus)

File konfigurasi cloud yang sudah tidak diperlukan untuk self-hosting:

| File | Keterangan |
|------|------------|
| `render.yaml` | Konfigurasi Render (cloud) — dihapus |
| `zeabur.json` | Konfigurasi Zeabur (cloud) — dihapus |
| `koyeb.yaml` | Konfigurasi Koyeb (cloud) — dihapus |
| `railway.toml` | Konfigurasi Railway (cloud) — dihapus |
| `.replit` | Konfigurasi Replit (cloud) — dihapus |
| `replit.nix` | Konfigurasi Replit (cloud) — dihapus |