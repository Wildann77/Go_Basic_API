# Go API Project

Project ini adalah REST API sederhana yang dibangun menggunakan **Go (Golang)**, **Gin Framework**, **GORM**, **PostgreSQL**, dan **Redis**.

## Fitur
- RESTful API menggunakan Gin
- ORM menggunakan GORM
- Database PostgreSQL
- Caching/Session menggunakan Redis
- Authentication (JWT)
- Dockerize project services
- Hot reload development (Air)..

## Prasyarat
- [Go](https://golang.org/dl/) (versi 1.21 atau terbaru)
- [Docker](https://www.docker.com/products/docker-desktop) & [Docker Compose](https://docs.docker.com/compose/install/)
- [Air](https://github.com/air-verse/air) (Optional, untuk hot reload)

## Cara Menjalankan

### 1. Persiapan Environment
Copy file `.env.example` menjadi `.env` dan sesuaikan konfigurasinya:
```bash
cp .env.example .env
```

### 2. Setup Awal & Jalankan Services
Perintah ini akan mendownload dependencies dan menyalakan Docker services (DB & Redis):
```bash
make setup
```

### 3. Menjalankan Aplikasi
Untuk development dengan hot reload:
```bash
make dev
```
Atau tanpa hot reload:
```bash
make run
```

## Perintah Makefile yang Tersedia
- `make setup`: Setup awal (deps + services-up).
- `make dev`: Menjalankan aplikasi dengan hot reload (Air).
- `make services-up`: Menyalakan PostgreSQL, Redis, dan Adminer.
- `make services-down`: Mematikan Docker containers.
- `make services-logs`: Melihat log dari services.
- `make build`: Membuat binary di folder `bin/`.
- `make clean`: Menghapus file binary.
- `make test`: Menjalankan unit test.
- `make migrate-up`: Menjalankan migrasi database.

## Struktur Folder
- `cmd/api/`: Entry point aplikasi (main.go).
- `internal/`: Logika bisnis internal (models, repository, services, config, handlers).
- `pkg/`: Library atau utility yang bisa digunakan kembali.
- `docker-compose.yml`: Konfigurasi Docker untuk database dan redis.

