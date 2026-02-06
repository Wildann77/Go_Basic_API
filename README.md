# Go API Project

Project ini adalah REST API sederhana yang dibangun menggunakan **Go (Golang)**, **Gin Framework**, **GORM**, **PostgreSQL**, dan **Redis**.

## Fitur
- RESTful API menggunakan Gin
- ORM menggunakan GORM
- Database PostgreSQL
- Caching/Session menggunakan Redis
- Authentication (JWT)
- Dockerize project

## Prasyarat
- [Go](https://golang.org/dl/) (versi 1.21 atau terbaru)
- [Docker](https://www.docker.com/products/docker-desktop) & [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/) (opsional, untuk menjalankan perintah Makefile)

## Cara Menjalankan

### 1. Persiapan Environment
Copy file `.env.example` menjadi `.env` dan sesuaikan konfigurasinya:
```bash
cp .env.example .env
```

### 2. Menjalankan Infrastruktur (Database & Redis)
Gunakan Docker Compose untuk menjalankan PostgreSQL dan Redis:
```bash
docker-compose up -d
```

### 3. Menjalankan Aplikasi
Anda bisa menjalankan aplikasi langsung menggunakan Go atau melalui Makefile.

**Menggunakan Go:**
```bash
go run cmd/api/main.go
```

**Menggunakan Makefile:**
```bash
make run
```

## Perintah Makefile yang Tersedia
- `make build`: Membuat binary aplikasi.
- `make run`: Menjalankan aplikasi.
- `make test`: Menjalankan unit test.
- `make docker-up`: Menjalankan container (db & redis).
- `make docker-down`: Menghentikan container.
- `make clean`: Menghapus file binary.

## Struktur Folder
- `cmd/api/`: Entry point aplikasi (main.go).
- `internal/`: Logika bisnis internal (models, repository, services, config).
- `pkg/`: Library atau utility yang bisa digunakan kembali.
- `docker-compose.yml`: Konfigurasi Docker untuk database dan redis.
