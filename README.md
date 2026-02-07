# Go API Project

Project ini adalah REST API boilerplate yang dibangun menggunakan **Go (Golang)** dengan mengikuti prinsip **Clean Architecture**.

## 🚀 Fitur Utama
- **RESTful API**: Menggunakan [Gin Framework](https://github.com/gin-gonic/gin).
- **ORM & Database**: Menggunakan [GORM](https://gorm.io/) dengan PostgreSQL.
- **Authentication**: Keamanan menggunakan JWT (JSON Web Token).
- **Development**: Hot reload menggunakan [Air](https://github.com/air-verse/air).
- **Deployment**: Mendukung Docker & Docker Compose.
- **Middleware**: CORS, Logger, dan JWT Authentication.

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
2.  **Middleware**: Request diperiksa (Logger, CORS, atau JWT Auth jika rute terproteksi).
3.  **Handler**: Menerima request, validasi format JSON, lalu memanggil fungsi di **Service**.
4.  **Service**: Menjalankan logika bisnis (misal: hashing password, cek email duplikat), lalu memanggil **Repository**.
5.  **Repository**: Melakukan operasi ke **Database** menggunakan GORM.
6.  **Response**: Data dikembalikan dari Repo -> Service -> Handler, lalu Handler mengirim response JSON ke Client.

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
make services-up
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
| `make setup` | Install dependencies dan jalankan Docker services. |
| `make dev` | Jalankan server dengan hot reload (Air). |
| `make build` | Compile aplikasi menjadi binary. |
| `make test` | Jalankan unit testing. |
| `make services-up` | Jalankan PostgreSQL & Redis di background. |
| `make services-down` | Matikan semua services Docker. |

---
