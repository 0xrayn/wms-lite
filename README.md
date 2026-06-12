# WMS API

Inventory Management API dengan Go + Gin + PostgreSQL. Fokus utama project ini adalah penanganan stock transaction secara **atomic** menggunakan database transaction dan row-level locking, untuk mencegah race condition saat banyak request mengubah stok produk yang sama secara bersamaan.

## Fitur

- Auth dengan JWT (register & login, role admin/staff)
- CRUD produk (admin only untuk create)
- Transaksi stok IN/OUT dengan validasi stok dan database transaction
- Row-level locking (`SELECT ... FOR UPDATE`) untuk handle concurrency
- History transaksi per produk
- Laporan ringkasan stok (current stock, total IN, total OUT, nilai stok)

## Setup

1. Pastikan PostgreSQL sudah berjalan, buat database baru:
   ```bash
   createdb wms
   ```

2. Copy `.env.example` ke `.env` dan sesuaikan konfigurasinya:
   ```bash
   cp .env.example .env
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Jalankan server:
   ```bash
   go run main.go
   ```

   Saat start, aplikasi otomatis menjalankan migration (membuat tabel `users`, `products`, `stock_transactions` kalau belum ada). File `db/schema.sql` tetap disediakan sebagai referensi/dokumentasi schema, tapi tidak perlu dijalankan manual.

## Dokumentasi API

### Auth

**Register**
```
POST /auth/register
{
  "email": "admin@example.com",
  "password": "secret123",
  "role": "admin"
}
```

**Login**
```
POST /auth/login
{
  "email": "admin@example.com",
  "password": "secret123"
}
```
Response berisi `token` JWT yang dipakai di header `Authorization: Bearer <token>` untuk endpoint lainnya.

### Products

**Create product (admin only)**
```
POST /products
{
  "name": "Kabel HDMI 2m",
  "sku": "HDMI-2M-001",
  "price": 35000
}
```

**Get all products**
```
GET /products
```

### Transactions

**Create transaction (stock in/out)**
```
POST /transactions
{
  "product_id": 1,
  "type": "IN",
  "quantity": 50,
  "note": "Restock dari supplier A"
}
```

Kalau `type` adalah `OUT` dan quantity melebihi stok yang tersedia, request akan ditolak dengan status 400.

**Get transaction history per product**
```
GET /products/:id/transactions
```

### Reports

**Stock summary**
```
GET /reports/stock-summary
```
Mengembalikan daftar produk beserta current stock, total IN, total OUT, dan nilai stok (price × current_stock).

## Konsep teknis yang ditekankan

- **Database transaction**: setiap perubahan stok dan pencatatan history dilakukan dalam satu transaction, kalau salah satu gagal, semuanya rollback.
- **Row-level locking**: menggunakan `SELECT ... FOR UPDATE` saat membaca stok produk sebelum update, supaya request lain ke produk yang sama harus menunggu (mencegah dua transaksi mengubah stok secara bersamaan dan menghasilkan data yang salah).

## Pengembangan lanjutan (opsional)

- Tambah Dockerfile + docker-compose untuk app + Postgres
- Tambah GitHub Actions untuk run test otomatis
- Tambah pagination & filter di endpoint `/products` dan history transaksi
