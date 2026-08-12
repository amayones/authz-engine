# Panduan Deployment Gratis — Zeabur + Supabase

Panduan ini menjelaskan cara hosting `authz-engine` (aplikasi Go) **gratis** menggunakan:

1. **Supabase** — database PostgreSQL gratis (500MB)
2. **Zeabur** — platform hosting modern dari Asia, mendukung Go native, free tier

> **Mengapa Zeabur?**
> - **Region dekat**: Hong Kong / Tokyo — sangat dekat dengan Supabase region Singapore (latency rendah)
> - **Go native**: auto-detect `go.mod`, build & deploy otomatis
> - **Free tier**: 1 service gratis, tanpa kartu kredit
> - **Modern & simple**: deploy dari GitHub dengan satu klik, auto HTTPS
> - **Tanpa konflik konfigurasi** seperti Replit

---

## Ringkasan

| Komponen | Platform | Gratis | Tanpa Kartu Kredit |
|----------|----------|--------|-------------------|
| Database PostgreSQL | [Supabase](https://supabase.com) | 500MB | ✅ |
| Hosting App Go | [Zeabur](https://zeabur.com) | Free tier | ✅ (via GitHub) |

---

## Langkah 1: Persiapkan Proyek di GitHub

```bash
cd c:\Apache2462\htdocs\authz-engine
git add .
git commit -m "tambah konfigurasi Zeabur + Supabase"
git push origin main
```

---

## Langkah 2: Buat Database PostgreSQL di Supabase

1. Buka https://supabase.com → **Start your project** (login GitHub/email)
2. **New project**:
   - **Name**: `authz-engine`
   - **Password**: buat password kuat, simpan
   - **Region**: **Southeast Asia (Singapore)**
3. Setelah jadi (±2 menit), buka **Settings → Database → Connection string**
4. Pilih **URI**, salin:
   ```
   postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

---

## Langkah 3: Deploy ke Zeabur

### 3.1 Daftar Zeabur

1. Buka https://zeabur.com
2. Klik **Sign up** → **Continue with GitHub** (bukan kartu kredit)
3. Zeabur otomatis memberi free tier

### 3.2 Create Project & Deploy

1. Klik **Create Project** → beri nama (mis. `authz-engine`)
2. Klik **Deploy Service** → pilih **GitHub**
3. Hubungkan repo `amayones/authz-engine`
4. Zeabur otomatis mendeteksi **Dockerfile** di repo dan build
5. Setelah build selesai, klik **Generate Domain** untuk dapat URL:
   ```
   https://authz-engine-xxxx.zeabur.app
   ```

> Konfigurasi `zeabur.json` sudah ada di repo, jadi Zeabur otomatis paham:
> - Build: `dockerfile`
> - Port: `8080`

### 3.3 Set Environment Variables

1. Di dashboard service, buka tab **Variables**
2. Tambahkan:

   | Key | Value |
   |-----|-------|
   | `AUTHZ_DB_DRIVER` | `postgres` |
   | `AUTHZ_DB_CONN` | *(connection string Supabase dari Langkah 2)* |
   | `AUTHZ_ADDR` | `:8080` |
   | `AUTHZ_AUTO_MIGRATE` | `true` |

3. Zeabur otomatis redeploy dengan env baru

### 3.4 Verifikasi

```bash
curl https://authz-engine-xxxx.zeabur.app/health
```

Response:
```json
{"status":"ok"}
```

---

## Langkah 4: Buat API Key (di komputer lokal)

```bash
cd c:\Apache2462\htdocs\authz-engine

go run ./cmd/genkey \
  -driver postgres \
  -db "postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres" \
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

### 5.1 Buat role
```bash
curl -X POST https://authz-engine-xxxx.zeabur.app/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"name": "editor", "permissions": ["invoice:read", "invoice:write"]}'
```

### 5.2 Assign role
```bash
curl -X POST https://authz-engine-xxxx.zeabur.app/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin
```bash
curl -X POST https://authz-engine-xxxx.zeabur.app/can \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "resource": "invoice", "action": "read"}'
```

Response:
```json
{"allowed":true}
```

---

## Deploy Ulang Otomatis

Zeabur otomatis **redeploy** setiap push ke branch `main` di GitHub.

```bash
git push origin main
```

---

## Troubleshooting

### Error: `connection refused` saat migration

- Cek `AUTHZ_DB_CONN` sudah benar di Variables
- Jika password mengandung `@`/`:`, gunakan URL-encoding (`%40`, `%3A`)
- Gunakan **URI** (bukan Transaction/Pooler) di Supabase

### App tidak bisa diakses

1. Cek tab **Logs** di dashboard Zeabur
2. Pastikan **Port 8080** sudah benar (dari `zeabur.json`)
3. Pastikan `AUTHZ_ADDR=:8080` di env vars
4. Cek healthcheck `/health` di browser

### Migration gagal di log

- Pastikan `AUTHZ_AUTO_MIGRATE=true` sudah di-set
- Pastikan `AUTHZ_DB_DRIVER=postgres` (bukan sqlserver)

---

## Catatan Penting

- **Zeabur Free tier** — 1 service gratis, tanpa kartu kredit
- **Supabase** gratis 500MB
- **Region dekat**: Zeabur Hong Kong/Tokyo + Supabase Singapore = latency rendah
- **Auto-deploy** dari GitHub
- **HTTPS otomatis** via `*.zeabur.app`

---

## Alternatif Lain (Juga Gratis)

| Kombinasi | Database | App | Gratis | Catatan |
|-----------|----------|-----|--------|---------|
| Supabase + **Zeabur** | [Supabase](https://supabase.com) | [Zeabur](https://zeabur.com) | ✅ Free tier | **Rekomendasi terbaik** (region dekat) |
| Supabase + **Koyeb** | [Supabase](https://supabase.com) | [Koyeb](https://koyeb.com) | ✅ $5/bulan refresh | Region jauh (Paris/NY) |
| Supabase + **Fly.io** | [Supabase](https://supabase.com) | [Fly.io](https://fly.io) | ✅ (2-3 VM gratis) | Perlu kartu kredit verifikasi |
| Supabase + **Replit** | [Supabase](https://supabase.com) | [Replit](https://replit.com) | ✅ | Idle & konflik nix |
| Supabase + **Railway** | [Supabase](https://supabase.com) | [Railway](https://railway.app) | ⚠️ Trial habis | $5 trial sekali |