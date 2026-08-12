# Panduan Self-Hosting — Jalankan dari PC/Laptop (SQL Server)

Panduan ini menjelaskan cara menjalankan `authz-engine` **langsung dari PC/laptop Anda** (self-host) menggunakan **SQL Server** yang sudah terinstall di laptop Anda.

---

## Ringkasan

| Komponen | Cara |
|----------|------|
| Aplikasi Go | Jalankan langsung dari PC/laptop (Windows) |
| Database | **SQL Server** (sudah terinstall di laptop) |
| Akses dari luar | Port forwarding / tunnel (ngrok, cloudflared) |

---

## Langkah 1: Siapkan Database di SQL Server

### 1.1 Buat Database

Buka **SQL Server Management Studio (SSMS)** atau **Azure Data Studio**, lalu jalankan:

```sql
CREATE DATABASE authzdb;
```

### 1.2 Connection String

```
sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
```

> **User**: `may` | **Password**: `may`

---

## Langkah 2: Build Aplikasi

Buka terminal di folder proyek:

```bash
cd c:\Apache2462\htdocs\authz-engine
go build -o authz-server.exe ./cmd/server
go build -o authz-genkey.exe ./cmd/genkey
```

---

## Langkah 3: Jalankan Server

### Cara 1: Pakai `start.bat` (paling mudah)

1. Double-click `start.bat` (sudah dikonfigurasi dengan user `may` / password `may`)

### Cara 2: Manual (Command Prompt / PowerShell)

```cmd
set AUTHZ_DB_DRIVER=sqlserver
set AUTHZ_DB_CONN=sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true
set AUTHZ_ADDR=:8080
set AUTHZ_AUTO_MIGRATE=true
authz-server.exe
```

### Cara 3: Manual (Git Bash)

```bash
export AUTHZ_DB_DRIVER=sqlserver
export AUTHZ_DB_CONN='sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true'
export AUTHZ_ADDR=':8080'
export AUTHZ_AUTO_MIGRATE='true'
./authz-server.exe
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

## Langkah 4: Buat API Key

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

> **PENTING**: Simpan API key — hanya ditampilkan sekali.

---

## Langkah 5: Uji API

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
title authz-engine server
cd /d %~dp0

REM Database: SQL Server lokal
set AUTHZ_DB_DRIVER=sqlserver
set AUTHZ_DB_CONN=sqlserver://may:may@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true

REM HTTP server
set AUTHZ_ADDR=:8080

REM Auto migration saat startup
set AUTHZ_AUTO_MIGRATE=true

authz-server.exe
pause
```

### `stop.bat` — Stop server

```bat
@echo off
taskkill /f /im authz-server.exe
```

---

## Troubleshooting

### Error: `connection refused` saat migration

- Pastikan SQL Server sudah berjalan (Service: `SQL Server (MSSQLSERVER)`)
- Cek connection string benar (user `may`, password `may`, port 1433)
- Pastikan database `authzdb` sudah dibuat

### Error: `login failed for user 'may'`

- Pastikan password `may` benar
- Pastikan SQL Server menggunakan **SQL Server Authentication** (bukan Windows Auth)
- Di SSMS: Properties → Security → pilih **SQL Server and Windows Authentication mode**

### Error: `database "authzdb" does not exist`

- Jalankan di SSMS: `CREATE DATABASE authzdb;`

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
- **SQL Server** = data tersimpan di laptop Anda
- **Tunnel** (ngrok/cloudflared) = akses dari internet tanpa port forwarding manual
- **Auto-start** = aplikasi jalan otomatis saat Windows boot

---

## File yang Tidak Diperlukan (Sudah Dihapus)

File konfigurasi cloud dan PostgreSQL yang sudah tidak diperlukan:

| File | Keterangan |
|------|------------|
| `internal/store/postgres/` | Store PostgreSQL — dihapus |
| `migrations/postgres/` | Migration PostgreSQL — dihapus |
| `render.yaml` | Konfigurasi Render (cloud) — dihapus |
| `zeabur.json` | Konfigurasi Zeabur (cloud) — dihapus |
| `koyeb.yaml` | Konfigurasi Koyeb (cloud) — dihapus |
| `railway.toml` | Konfigurasi Railway (cloud) — dihapus |
| `.replit` | Konfigurasi Replit (cloud) — dihapus |
| `replit.nix` | Konfigurasi Replit (cloud) — dihapus |