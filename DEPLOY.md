# Panduan Deployment Gratis — Railway + Supabase

Panduan ini menjelaskan cara hosting `authz-engine` (aplikasi Go) **gratis** menggunakan:

1. **Supabase** — database PostgreSQL gratis (500MB)
2. **Railway** — hosting Go modern, simple, dan stabil

> **Mengapa Railway bukan Replit/Render?**
> Railway lebih modern dan simple untuk aplikasi Go: auto-detect `go.mod`, deploy dari GitHub dengan satu klik, HTTPS otomatis, dan tanpa konflik konfigurasi seperti Replit.

---

## Ringkasan

| Komponen | Platform | Gratis | Tanpa Kartu Kredit |
|----------|----------|--------|-------------------|
| Database PostgreSQL | [Supabase](https://supabase.com) | 500MB | ✅ |
| Hosting App Go | [Railway](https://railway.app) | $5 kredit/bulan | ✅ (via GitHub) |

> **Catatan biaya**: Railway memberi **$5 kredit gratis per bulan** tanpa kartu kredit. Satu service Go berjalan 24/7 dengan RAM 512MB kira-kira menghabiskan **$3-4/bulan** — cukup untuk 1 service ringan secara gratis terus-menerus.

---

## Langkah 1: Persiapkan Proyek di GitHub

Push proyek ke GitHub:

```bash
cd c:\Apache2462\htdocs\authz-engine
git add .
git commit -m "tambah konfigurasi Railway + optimasi Docker"
git push origin main
```

> Jika repo belum ada, buat di https://github.com/new lalu hubungkan.

---

## Langkah 2: Buat Database PostgreSQL di Supabase

1. Buka https://supabase.com → **Start your project** (login GitHub/email)
2. Klik **New project**:
   - **Name**: `authz-engine`
   - **Database Password**: buat password kuat, simpan
   - **Region**: **Southeast Asia (Singapore)**
3. Setelah dibuat (±2 menit), buka **Settings → Database → Connection string**
4. Pilih **URI**, salin — formatnya:
   ```
   postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

---

## Langkah 3: Deploy ke Railway

### 3.1 Daftar Railway

1. Buka https://railway.app
2. Klik **Sign up** → **Continue with GitHub**
3. Grant akses GitHub yang diminta (bukan kartu kredit)

### 3.2 Buat Project dari Repository

1. Klik **New Project**
2. Pilih **Deploy from GitHub repo**
3. Pilih repo `amayones/authz-engine`
4. Railway otomatis mendeteksi Go dan men-deploy

> Konfigurasi (`railway.toml`) sudah ada di repo, jadi Railway otomatis tahu:
> - Build: `go build -o authz-server ./cmd/server`
> - Start: `./authz-server`
> - Healthcheck: `/health`

### 3.3 Set Environment Variables (Variables)

1. Di dashboard project Railway, klik service `authz-engine`
2. Buka tab **Variables**
3. Tambahkan:

   | Key | Value |
   |-----|-------|
   | `AUTHZ_DB_DRIVER` | `postgres` |
   | `AUTHZ_DB_CONN` | *(connection string Supabase dari Langkah 2)* |
   | `AUTHZ_ADDR` | `:8080` |
   | `AUTHZ_AUTO_MIGRATE` | `true` |

4. Railway akan otomatis redeploy dengan env baru

### 3.4 Dapatkan URL

1. Buka tab **Settings** pada service
2. Bagian **Networking** → **Generate Domain** → klik tombol generate
3. Railway membuat URL seperti: `https://authz-engine-production.up.railway.app`

### 3.5 Verifikasi

Buka di browser:

```
https://authz-engine-production.up.railway.app/health
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
curl -X POST https://authz-engine-production.up.railway.app/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "editor",
    "permissions": ["invoice:read", "invoice:write"]
  }'
```

### 5.2 Assign role ke user

```bash
curl -X POST https://authz-engine-production.up.railway.app/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin (Can)

```bash
curl -X POST https://authz-engine-production.up.railway.app/can \
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

Railway otomatis **redeploy** setiap kali ada push ke `main` di GitHub.
Tidak perlu melakukan apa-apa — cukup push:

```bash
git push origin main
```

---

## Deployment via Dockerfile (opsional)

Jika ingin memakai Dockerfile (bukan Nixpacks), ubah di `railway.toml`:

```toml
[build]
builder = "DOCKERFILE"
```

Dockerfile sudah dioptimasi:

- Build stage menggunakan `golang:1.26-alpine` (ringan)
- Runtime stage hanya `alpine:3.20` (image sangat kecil ~15MB)
- Migrations disertakan di image

---

## Troubleshooting

### Error: `connection refused` saat migration

- Pastikan connection string Supabase benar
- Jika password mengandung `@` atau `:`, gunakan **URL-encoding** (mis. `%40` untuk `@`, `%3A` untuk `:`)
- Pastikan memilih **URI** (bukan Transaction/Pooler)

### Error: `role "postgres" does not exist`

- Di Supabase, gunakan **URI** (bukan Transaction pooler)

### App tidak bisa diakses

1. Cek tab **Deployments** di Railway — apakah deployment sukses?
2. Cek tab **Logs** untuk error
3. Pastikan `AUTHZ_ADDR=:8080` dan **Generate Domain** sudah dilakukan
4. Pastikan healthcheck `/health` merespons

### Migration gagal muncul di log

- Pastikan `AUTHZ_AUTO_MIGRATE=true` di Variables
- Pastikan `AUTHZ_DB_CONN` benar dan dapat diakses dari Railway (Supabase memungkinkan koneksi dari cloud)

### Kredit $5 habis

- Railway memberi $5 kredit/bulan. Satu service Go 512MB memakan ±$3-4/bulan
- Jika habis, cukup tunggu bulan berikutnya, atau upgrade kapan saja

---

## Catatan Penting

- **Supabase** gratis 500MB — cukup untuk aplikasi kecil
- **Railway** memberi $5 kredit gratis/bulan — cukup untuk 1 service Go ringan
- **Migration otomatis** berjalan saat app start (`AUTHZ_AUTO_MIGRATE=true`)
- **Auto-deploy** dari GitHub setiap push ke `main`
- **HTTPS otomatis** via Generate Domain di Railway
- **Tidak ada downtime idle** seperti Replit — Railway berjalan 24/7

---

## Alternatif Lain (Jika Railway Tidak Cocok)

| Kombinasi | Database | App | Catatan |
|-----------|----------|-----|---------|
| Supabase + **Fly.io** | [Supabase](https://supabase.com) | [Fly.io](https://fly.io) | Perlu kartu kredit untuk verifikasi |
| Supabase + **Render** | [Supabase](https://supabase.com) | [Render](https://render.com) | Perlu kartu kredit untuk verifikasi |
| Supabase + **Replit** | [Supabase](https://supabase.com) | [Replit](https://replit.com) | Gratis, tapi idle & konflik konfigurasi |
| Supabase + **Railway** | [Supabase](https://supabase.com) | [Railway](https://railway.app) | **Terbaik untuk Go + Supabase** |

> **Rekomendasi**: Untuk aplikasi Go + Supabase, **Railway** adalah pilihan paling modern, simple, dan stabil — tanpa kartu kredit, auto-deploy, HTTPS otomatis, dan tanpa downtime idle.