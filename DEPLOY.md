# Panduan Deployment Gratis Selamanya — Koyeb + Supabase

Panduan ini menjelaskan cara hosting `authz-engine` (aplikasi Go) **gratis selamanya** (bukan trial) menggunakan:

1. **Supabase** — database PostgreSQL gratis (500MB)
2. **Koyeb** — hosting modern yang memberi $5 kredit/bulan **di-refresh terus-menerus** (bukan trial)

> **Kok bukan Replit/Railway/Render?**
> - Replit: idle & konflik konfigurasi (nix env)
> - Railway: $5 hanya trial, kredit habis setelah trial berakhir
> - Render: perlu kartu kredit untuk verifikasi
> - **Koyeb**: $5 kredit/bulan di-refresh terus, tanpa kartu kredit — benar-benar gratis selamanya

---

## Ringkasan

| Komponen | Platform | Gratis | Tanpa Kartu Kredit |
|----------|----------|--------|-------------------|
| Database PostgreSQL | [Supabase](https://supabase.com) | 500MB | ✅ |
| Hosting App Go | [Koyeb](https://koyeb.com) | $5/bulan (di-refresh) | ✅ (via GitHub) |

> **Koyeb Free tier**: `instance_types: free` — 1 service gratis, RAM 512MB, bandwidth 100GB/bulan. Kredit $5 di-refresh setiap bulan, jadi **tidak akan pernah habis** selama Anda stay di free tier.

---

## Langkah 1: Persiapkan Proyek di GitHub

```bash
cd c:\Apache2462\htdocs\authz-engine
git add .
git commit -m "tambah konfigurasi Koyeb + Supabase"
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

## Langkah 3: Deploy ke Koyeb

### 3.1 Daftar Koyeb

1. Buka https://koyeb.com
2. Klik **Sign up** → **Continue with GitHub** (bukan kartu kredit)
3. Koyeb otomatis memberikan $5 kredit/bulan yang **di-refresh terus**

### 3.2 Create App dari GitHub

1. Klik **Create App**
2. Pilih tab **GitHub**, hubungkan repo `amayones/authz-engine`
3. Koyeb otomatis mendeteksi **Dockerfile** di repo
4. Atur:
   - **Region**: pilih **Singapore (sin)** — dekat Supabase region Singapore
   - **Instance Type**: `Free` (sudah default dari `koyeb.yaml`)
   - **Service Type**: `Web Service` (default)
   - **Port**: `8080` (sudah dari `koyeb.yaml`)
5. Klik **Deploy**

> Konfigurasi `koyeb.yaml` sudah ada di repo, jadi Koyeb otomatis paham:
> - Port: `8080`
> - Healthcheck: `/health`
> - Env vars: `AUTHZ_DB_DRIVER`, `AUTHZ_ADDR`, `AUTHZ_AUTO_MIGRATE`

### 3.3 Set Environment Variables

1. Setelah deploy pertama sukses, buka service → tab **Environment Variables**
2. Tambahkan `AUTHZ_DB_CONN`:
   ```
   key:   AUTHZ_DB_CONN
   value: postgresql://postgres.XXXXX:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```
3. Klik **Redeploy**

### 3.4 Dapatkan URL

Koyeb otomatis memberi URL:
```
https://authz-engine-<nama>.koyeb.app
```

### 3.5 Verifikasi

```bash
curl https://authz-engine-<nama>.koyeb.app/health
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
curl -X POST https://authz-engine-<nama>.koyeb.app/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"name": "editor", "permissions": ["invoice:read", "invoice:write"]}'
```

### 5.2 Assign role
```bash
curl -X POST https://authz-engine-<nama>.koyeb.app/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin
```bash
curl -X POST https://authz-engine-<nama>.koyeb.app/can \
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

Koyeb otomatis **redeploy** setiap push ke branch `main` di GitHub.

```bash
git push origin main
```

---

## Troubleshooting

### Error: `connection refused` saat migration

- Cek `AUTHZ_DB_CONN` sudah benar di Environment Variables
- Jika password mengandung `@`/`:`, gunakan URL-encoding (`%40`, `%3A`)
- Gunakan **URI** (bukan Transaction/Pooler) di Supabase

### App tidak bisa diakses

1. Cek tab **Logs** di dashboard Koyeb
2. Pastikan **Port 8080** sudah benar di service config
3. Pastikan `AUTHZ_ADDR=:8080` di env vars
4. Cek healthcheck `/health` di browser

### Migration gagal di log

- Pastikan `AUTHZ_AUTO_MIGRATE=true` sudah di-set
- Pastikan `AUTHZ_DB_DRIVER=postgres` (bukan sqlserver)

---

## Catatan Penting

- **Koyeb Free tier** — kredit $5/bulan di-refresh terus, bukan trial
- **Supabase** gratis 500MB
- **Tidak ada downtime idle** seperti Replit — Koyeb berjalan 24/7
- **Auto-deploy** dari GitHub
- **HTTPS otomatis** via `*.koyeb.app`

---

## Alternatif Lain (Juga Gratis Selamanya)

| Kombinasi | Database | App | Gratis Selamanya | Catatan |
|-----------|----------|-----|-------------------|---------|
| Supabase + **Koyeb** | [Supabase](https://supabase.com) | [Koyeb](https://koyeb.com) | ✅ $5/bulan refresh | **Rekomendasi terbaik** |
| Supabase + **Fly.io** | [Supabase](https://supabase.com) | [Fly.io](https://fly.io) | ✅ (2-3 VM gratis) | Perlu kartu kredit verifikasi |
| Supabase + **Replit** | [Supabase](https://supabase.com) | [Replit](https://replit.com) | ✅ | Idle & konflik nix |
| Supabase + **Render** | [Supabase](https://supabase.com) | [Render](https://render.com) | ⚠️ Perlu kartu | Verifikasi kartu kredit |
| Supabase + **Railway** | [Supabase](https://supabase.com) | [Railway](https://railway.app) | ⚠️ Trial habis | $5 trial sekali |