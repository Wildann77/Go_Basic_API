# Go API Project

Project ini adalah REST API boilerplate yang dibangun menggunakan **Go (Golang)** dengan mengikuti prinsip **Clean Architecture**.

## 🚀 Fitur Utama
- **RESTful API**: Menggunakan [Gin Framework](https://github.com/gin-gonic/gin).
- **ORM & Database**: Menggunakan [GORM](https://gorm.io/) dengan PostgreSQL.
- **Authentication**: Keamanan menggunakan JWT (JSON Web Token).
- **Rate Limiting**: Pembatasan request berbasis IP menggunakan Redis untuk perlindungan bruteforce & abuse.
- **Cache & Storage**: Integrasi Redis untuk rate limiting dan data caching (Cache-Aside Pattern).
- **Development**: Hot reload menggunakan [Air](https://github.com/air-verse/air).
- **Deployment**: Mendukung Docker & Docker Compose.
- **Middleware**: CORS, Logger, Rate Limiter, dan JWT Authentication.
- **ACID Transactions**: Menggunakan Context propagation untuk operasi atomik yang aman.

---

## 🏗️ Struktur Folder & Arsitektur
Proyek ini menggunakan pemisahan layer untuk memastikan kode mudah di-maintain:

- `cmd/api/`: Entry point utama aplikasi. Tempat perakitan (wiring) semua komponen.
- `internal/config/`: Konfigurasi environment dan inisialisasi database.
- `internal/handlers/`: Layer terluar (HTTP). Menangani request dan response.
- `internal/services/`: Layer logika bisnis (Core Logic).
- `internal/repository/`: Layer data akses. Berinteraksi langsung dengan Database.
- `internal/models/`: Definisi struct data dan schema database.
- `internal/middleware/`: Fungsi penghalang (Interceptor) seperti auth dan logger.

---

## 🔄 Alur Data (Data Flow)
Berikut adalah gambaran bagaimana sebuah request diproses:

1.  **Client**: Mengirim request ke endpoint (misal: `POST /api/v1/register`).
2.  **Middleware**: Request diperiksa oleh beberapa layer:
    - **CORS & Logger**: Menangani request origin dan pencatatan log.
    - **Rate Limiter**: Memastikan client tidak melebihi batas quota request (Redis-backed).
    - **JWT Auth**: Verifikasi token untuk rute yang membutuhkan akses login.
3.  **Handler**: Menerima request, validasi format JSON, lalu memanggil fungsi di **Service**.
4.  **Service**:
    - **Caching Check**: Mencari data di Redis terlebih dahulu (Cache Hit).
    - **Business Logic**: Jika tidak ada di cache (Cache Miss), lanjut ke logika bisnis dan panggil Repository.
5.  **Repository**: Melakukan operasi ke **Database** (PostgreSQL) jika data belum terkecached.
6.  **Response**: Data dikembalikan ke Service (disimpan ke cache jika baru diambil dari DB) -> Handler -> Client.

---

## 🔐 Pola Transaksi (ACID)
Proyek ini mendukung **Database Transactions** penuh untuk menjaga integritas data (Atomicity, Consistency, Isolation, Durability).
- **Context-Based**: Transaksi diteruskan secara implisit melalui `context.Context`.
- **Atomic Service**: Logika bisnis kompleks di Service Layer dapat dibungkus dalam satu transaksi.
- **Repository Agnostic**: Repository secara otomatis mendeteksi apakah sedang berada dalam transaksi atau tidak.
---

## 🛡️ Rate Limiting
Proyek ini menggunakan **Distributed Rate Limiting** menggunakan Redis untuk memastikan performa tetap terjaga dan aman dari serangan DDoS ringan atau Brute-force.
- **Global Limit**: Membatasi semua request masuk (default: 100 req/menit).
- **Strict Limit**: Diterapkan pada rute sensitif seperti `/login` dan `/register` (default: 5 req/menit).
- **Redis-Backed**: Quota request tersimpan secara terpusat di Redis, memungkinkan skalabilitas horizontal (multi-instance).
---

## ⚡ Caching Strategy
Proyek ini mengimplementasikan **Cache-Aside Pattern** untuk meningkatkan performa read-heavy operations:
- **Hit**: Data diambil langsung dari Redis (sangat cepat).
- **Miss**: Data diambil dari DB, lalu disimpan ke Redis dengan TTL (Time To Live).
- **Invalidation**: Cache otomatis dihapus saat terjadi update/delete data (Data Consistency).
---

## 🛠️ Persyaratan Sistem
- **Go**: 1.23+
- **Docker & Docker Compose**
- **Air** (Opsional, untuk hot reload): `go install github.com/air-verse/air@latest`

---

## 🏎️ Cara Menjalankan

### 1. Inisialisasi Environment
Salin file contoh env dan sesuaikan jika diperlukan:
```bash
cp .env.example .env
```

### 2. Jalankan Infrastruktur (DB & Redis)
```bash
make up
```

### 3. Jalankan Aplikasi
**Mode Development (Hot Reload):**
```bash
make dev
```
**Mode Standar:**
```bash
make run
```

---

## 📜 Perintah Makefile
| Perintah | Deskripsi |
| :--- | :--- |
| `make setup` | Setup awal (download dependencies & jalankan services). |
| `make up` | Jalankan PostgreSQL & Redis menggunakan Docker. |
| `make down` | Hentikan semua services Docker. |
| `make status` | Cek status services Docker yang berjalan. |
| `make logs` | Intip logs dari services Docker. |
| `make dev` | Jalankan aplikasi dengan Hot Reload (Air). |
| `make run` | Jalankan aplikasi secara langsung (Go run). |
| `make build` | Compile aplikasi menjadi binary di folder `bin/`. |
| `make test` | Jalankan semua unit tests. |
| `make deps` | Download dan merapikan Go modules. |
| `make migrate-up` | Jalankan database migrations. |
| `make migrate-down` | Rollback database migrations. |
| `make clean` | Hapus folder `bin/`. |

---
