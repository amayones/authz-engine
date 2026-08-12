# Panduan Deployment — Render (App) + Supabase (Database)

Panduan ini menjelaskan cara hosting `authz-engine` (aplikasi Go) menggunakan:

1. **Render** — hosting aplikasi Go (free tier, 750 jam/bulan)
2. **Supabase** — database PostgreSQL gratis (500MB)

> **Kenapa database pakai Supabase, bukan Render?**
> Database PostgreSQL di Render **berbayar** (mulai $7/bulan). Supabase memberi **500MB gratis** — jauh lebih hemat. Jadi: **Render untuk deploy app, Supabase untuk database**.

---

## Ringkasan

| Komponen | Platform | Biaya | Kartu Kredit |
|----------|----------|-------|--------------|
| Hosting App Go | [Render](https://render.com) | Free tier (750 jam/bulan) | ✅ Perlu untuk verifikasi |
| Database PostgreSQL | [Supabase](https://supabase.com) | Gratis 500MB | ✅ Tidak |

> **Catatan**: Render memerlukan **kartu kredit untuk verifikasi** saat daftar, tapi **tidak akan ditagih** selama Anda stay di free tier. Ini syarat umum platform cloud modern.

---

## Langkah 1: Persiapkan Proyek di GitHub

```bash
cd c:\Apache2462\htdocs\authz-engine
git add .
git commit -m "tambah konfigurasi Render + Supabase"
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

## Langkah 3: Deploy ke Render

### 3.1 Daftar Render

1. Buka https://render.com
2. Klik **Sign up** → **Continue with GitHub**
3. Verifikasi kartu kredit (tidak akan ditagih di free tier)

### 3.2 Deploy via Blueprint

1. Klik **New +** → **Blueprint**
2. Pilih repo `amayones/authz-engine`
3. Render otomatis membaca `render.yaml` dan membuat service:
   - **Runtime**: Go
   - **Build**: `go build -o authz-server ./cmd/server`
   - **Start**: `./authz-server`
   - **Healthcheck**: `/health`
   - **Plan**: Free

### 3.3 Set Environment Variables

1. Di dashboard service, buka tab **Environment**
2. Tambahkan `AUTHZ_DB_CONN`:
   ```
   key:   AUTHZ_DB_CONN
   value: postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```
3. Render otomatis redeploy

### 3.4 Dapatkan URL

Render otomatis memberi URL:
```
https://authz-engine.onrender.com
```

### 3.5 Verifikasi

```bash
curl https://authz-engine.onrender.com/health
```

Response:
```json
{"status":"ok"}
```

> **Catatan**: Render free tier akan **sleep** setelah 15 menit tidak aktif. Saat diakses lagi, akan restart otomatis (butuh ±30 detik pertama kali).

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
curl -X POST https://authz-engine.onrender.com/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"name": "editor", "permissions": ["invoice:read", "invoice:write"]}'
```

### 5.2 Assign role
```bash
curl -X POST https://authz-engine.onrender.com/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin
```bash
curl -X POST https://authz-engine.onrender.com/can \
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

Render otomatis **redeploy** setiap push ke branch `main` di GitHub.

```bash
git push origin main
```

---

## Troubleshooting

### Error: `connection refused` saat migration

- Cek `AUTHZ_DB_CONN` sudah benar di Environment
- Jika password mengandung `@`/`:`, gunakan URL-encoding (`%40`, `%3A`)
- Gunakan **URI** (bukan Transaction/Pooler) di Supabase

### App tidak bisa diakses

1. Cek tab **Logs** di dashboard Render
2. Pastikan **Port 10000** sudah benar (Render free tier pakai port 10000)
3. Pastikan `AUTHZ_ADDR=:10000` di env vars
4. Cek healthcheck `/health` di browser

### Migration gagal di log

- Pastikan `AUTHZ_AUTO_MIGRATE=true` sudah di-set
- Pastikan `AUTHZ_DB_DRIVER=postgres` (bukan sqlserver)

### App sleep (idle)

- Render free tier sleep setelah 15 menit tidak aktif
- Saat diakses lagi, restart otomatis (±30 detik)
- Ini normal untuk free tier

---

## Catatan Penting

- **Render free tier** — 750 jam/bulan, sleep saat idle
- **Supabase** gratis 500MB
- **Database TIDAK dibuat di Render** (mahal) — pakai Supabase
- **Auto-deploy** dari GitHub
- **HTTPS otomatis** via `*.onrender.com`

---

## Alternatif Lain

| Kombinasi | Database | App | Catatan |
|-----------|----------|-----|---------|
| Supabase + **Render** | [Supabase](https://supabase.com) | [Render](https://render.com) | **Pilihan Anda** — stabil, perlu kartu verifikasi |
| Supabase + **Zeabur** | [Supabase](https://supabase.com) | [Zeabur](https://zeabur.com) | Region Asia, tanpa kartu |
| Supabase + **Koyeb** | [Supabase](https://supabase.com) | [Koyeb](https://koyeb.com) | $5/bulan refresh |
| Supabase + **Fly.io** | [Supabase](https://supabase.com) | [Fly.io](https://fly.io) | Perlu kartu verifikasi |