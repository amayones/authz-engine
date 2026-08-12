# Panduan Deployment Gratis Tanpa Kartu Kredit

Panduan ini menjelaskan cara hosting `authz-engine` secara **gratis tanpa kartu kredit** menggunakan:

1. **Supabase** — database PostgreSQL gratis (500MB)
2. **Adaptable.io** — hosting aplikasi Go gratis (auto-deploy dari GitHub)

---

## Ringkasan

| Komponen | Platform | Gratis | Tanpa Kartu Kredit |
|----------|----------|--------|-------------------|
| Database PostgreSQL | [Supabase](https://supabase.com) | 500MB | ✅ |
| Hosting App Go | [Adaptable.io](https://adaptable.io) | ✅ | ✅ |
| HTTPS Otomatis | Adaptable.io | ✅ | ✅ |

---

## Langkah 1: Persiapkan Proyek di GitHub

### 1.1 Push proyek ke GitHub

Buka terminal di folder proyek Anda:

```bash
cd c:\Apache2462\htdocs\authz-engine

# Tambahkan semua perubahan (termasuk dukungan PostgreSQL yang sudah ditambahkan)
git add .
git commit -m "tambah dukungan PostgreSQL + panduan deployment"
git push origin main
```

> Jika repo belum ada di GitHub, buat dulu di https://github.com/new lalu hubungkan:
> ```bash
> git remote add origin https://github.com/USERNAME/authz-engine.git
> git branch -M main
> git push -u origin main
> ```

---

## Langkah 2: Buat Database PostgreSQL Gratis di Supabase

### 2.1 Daftar Supabase

1. Buka https://supabase.com
2. Klik **Start your project**
3. Daftar dengan **GitHub** atau **Email** (bukan kartu kredit)

### 2.2 Buat Project

1. Setelah login, klik **New project**
2. Isi form:
   - **Name**: `authz-engine` (atau nama lain)
   - **Database Password**: buat password kuat, simpan baik-baik
   - **Region**: pilih **Southeast Asia (Singapore)** — terdekat dengan Indonesia
3. Klik **Create new project**

### 2.3 Ambil Connection String

1. Tunggu project selesai dibuat (±2 menit)
2. Di sidebar kiri, klik **Settings** → **Database**
3. Scroll ke bagian **Connection string**
4. Pilih **URI** (bukan Pooler)
5. Salin connection string, formatnya seperti ini:

   ```
   postgresql://postgres.XXXXX:YOUR-PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

   > **PENTING**: Ganti `YOUR-PASSWORD` dengan password database yang tadi Anda buat.
   > Simpan connection string ini — akan dipakai di Langkah 3.

---

## Langkah 3: Deploy Aplikasi ke Adaptable.io

### 3.1 Daftar Adaptable.io

1. Buka https://adaptable.io
2. Klik **Get Started** / **Sign Up**
3. Daftar dengan **GitHub** (bukan kartu kredit)

### 3.2 Deploy Repo

1. Setelah login, klik **New App**
2. Pilih **Deploy from existing repository**
3. Pilih repo GitHub Anda (`authz-engine`)
4. Pilih **main** branch

### 3.3 Konfigurasi Build

1. **Template**: pilih **Go**
2. **Build Command**: `go build -o authz-server ./cmd/server`
3. **Start Command**: `./authz-server`

### 3.4 Set Environment Variables

1. Di bagian **Environment Variables**, tambahkan:

   | Variabel | Nilai |
   |----------|-------|
   | `AUTHZ_DB_DRIVER` | `postgres` |
   | `AUTHZ_DB_CONN` | *(isi dengan connection string dari Supabase, Langkah 2.3)* |
   | `AUTHZ_ADDR` | `:8080` |
   | `AUTHZ_AUTO_MIGRATE` | `true` |

   Contoh:
   ```
   AUTHZ_DB_CONN = postgresql://postgres.abc123:password123@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

### 3.5 Deploy

1. Klik **Deploy App**
2. Tunggu proses build & deploy selesai (±2-5 menit)
3. Jika berhasil, akan muncul URL seperti: `https://authz-engine.adaptable.app`

### 3.6 Verifikasi

Buka URL aplikasi Anda di browser:

```
https://nama-app.adaptable.app/health
```

Response yang benar:

```json
{"status":"ok"}
```

> Jika muncul `{"error":"..."}` atau halaman error, cek tab **Deployments** / **Logs** di panel Adaptable untuk melihat error.

---

## Langkah 4: Buat API Key

API key dibuat dengan menjalankan `cmd/genkey` di komputer lokal. Anda perlu `go` terpasang.

### 4.1 Jalankan genkey

```bash
cd c:\Apache2462\htdocs\authz-engine

# Ganti <CONNECTION_STRING> dengan connection string Supabase Anda
go run ./cmd/genkey \
  -driver postgres \
  -db "postgresql://postgres.abc123:password123@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres" \
  -name "client1" \
  -rpm 120
```

Output:

```
API key berhasil dibuat. SIMPAN SEKARANG — tidak akan ditampilkan lagi:
ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```

> **PENTING**: Simpan API key ini dengan aman — hanya ditampilkan sekali.

---

## Langkah 5: Uji API

### 5.1 Buat role

```bash
curl -X POST https://nama-app.adaptable.app/roles \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "editor",
    "permissions": ["invoice:read", "invoice:write"]
  }'
```

### 5.2 Assign role ke user

```bash
curl -X POST https://nama-app.adaptable.app/roles/assign \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "role_name": "editor"}'
```

### 5.3 Cek izin (Can)

```bash
curl -X POST https://nama-app.adaptable.app/can \
  -H "X-API-Key: ak_9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "user:alice", "resource": "invoice", "action": "read"}'
```

Response:

```json
{"allowed":true}
```

---

## Troubleshooting

### Error: `connection refused` saat migration

Pastikan connection string Supabase benar dan password tidak mengandung karakter khusus yang perlu di-encode. Jika password mengandung `@` atau `:` gunakan URL-encoding.

### Error: `role "postgres" does not exist`

Pastikan Anda memilih **URI** (bukan **Transaction** / **Pooler**) di Supabase. Untuk production, bisa gunakan **Transaction pooler**.

### App mati (idle) di Adaptable

Adaptable.io mematikan app gratis setelah tidak digunakan (idle). Saat diakses lagi, akan restart otomatis. Ini normal untuk tier gratis.

### Ingin auto-deploy setiap push?

Adaptable.io otomatis mendeploy ulang setiap kali Anda push ke branch yang dipilih (default: `main`).

---

## Catatan Penting

- **Supabase** gratis 500MB — cukup untuk aplikasi kecil. Bisa upgrade kapan saja.
- **Adaptable.io** gratis dengan limit waktu aktif — mati saat idle, nyala lagi saat diakses.
- **Migration otomatis** berjalan saat app start (`AUTHZ_AUTO_MIGRATE=true`). Setelah pertama kali, table sudah dibuat.
- Untuk development lokal, Anda bisa pakai PostgreSQL lokal atau Docker.

---

## Alternatif Lain (Juga Tanpa Kartu Kredit)

| Kombinasi | Database | App |
|-----------|----------|-----|
| Neon (512MB) + Adaptable.io | [Neon](https://neon.tech) | [Adaptable](https://adaptable.io) |
| Supabase (500MB) + Railway ($5 kredit) | [Supabase](https://supabase.com) | [Railway](https://railway.app) |
| Aiven (1GB) + Glitch | [Aiven](https://aiven.io) | [Glitch](https://glitch.com) |
| Supabase (500MB) + Replit | [Supabase](https://supabase.com) | [Replit](https://replit.com) |