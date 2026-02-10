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
- **ACID Transactions**: Menggunakan Context propagation untuk operasi atomik yang aman.
- **Observability**: Structured Logging (JSON), Request ID tracking, dan Health Check yang mendalam.
- **Error Handling**: Custom Recovery middleware untuk menangani panic dan mencatat log secara aman.

---

## 🏗️ Struktur Folder & Arsitektur
Proyek ini menggunakan pemisahan layer untuk memastikan kode mudah di-maintain:

- `cmd/api/`: Entry point utama aplikasi. Tempat perakitan (wiring) semua komponen.
- `internal/config/`: Konfigurasi environment dan inisialisasi database.
- `internal/handlers/`: Layer terluar (HTTP). Menangani request dan response.
- `internal/services/`: Layer logika bisnis (Core Logic).
- `internal/repository/`: Layer data akses. Berinteraksi langsung dengan Database.
- `internal/models/`: Definisi struct data dan schema database.
- `internal/middleware/`: Fungsi penghalang (Interceptor) seperti auth, logger, dan recovery.
- `pkg/logger/`: Library structured logging berbasis `log/slog`.
- `pkg/utils/`: Library pendukung untuk response dan database context.

---

## 🔄 Alur Data (Data Flow)
Berikut adalah gambaran bagaimana sebuah request diproses:

1.  **Client**: Mengirim request ke endpoint (misal: `POST /api/v1/register`).
2.  **Middleware**: Request diperiksa oleh beberapa layer:
    - **Request ID**: Setiap request diberikan ID unik (`X-Request-ID`) untuk tracking log.
    - **CORS & Logger**: Menangani request origin dan pencatatan log terstruktur (JSON).
    - **Custom Recovery**: Melindungi aplikasi dari panic dan melog stack trace secara otomatis.
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

## 🗂️ Database Indexes
Optimasi pencarian dan pengurutan data dilakukan menggunakan **Strategic Indexing** pada level database:
- **Unique Indexes**: Menjamin keunikan data sekaligus mempercepat lookup (Email, Username).
- **Search Optimization**: Index pada kolom `full_name` untuk pencarian user yang efisien.
- **Filtering Optimization**: Index pada kolom `active` untuk mempercepat filtering status.
- **Sorting Optimization**: Index descending pada `created_at` untuk mendukung sinkronisasi data terbaru dengan performa tinggi.
- **GORM Integrated**: Semua index dikelola langsung melalui struct tags di GORM models untuk kemudahan pemeliharaan schema.
---

## 🩺 Observability & Monitoring
Proyek ini dirancang agar mudah dimonitor di production:
- **Structured Logging**: Menggunakan `log/slog` dengan output JSON (Standar Cloud-Native). Log menyertakan `latency`, `request_id`, `ip`, dan `user_agent`.
- **Request Identification**: Mendukung `X-Request-ID` di header response dan logs, memungkinkan penelusuran satu request dari awal hingga akhir (Distributed Tracing ready).
- **Health Check**: Endpoint `/health` memantau kesehatan semua komponen:
  - **PostgreSQL**: Melakukan ping ke DB.
  - **Redis**: Melakukan ping ke Redis cluster.
- **Panic Protection**: Middleware recovery kustom menjamin aplikasi tetap hidup meski terjadi error fatal di goroutine, sambil mencatat detil stack trace ke log sistem.
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

Selamat Mencoba !!