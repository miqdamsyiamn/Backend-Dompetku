# DompetKu API

Backend API untuk aplikasi catatan keuangan pribadi menggunakan Go (Gin) dan MongoDB.

## Tech Stack

- **Go** - Programming Language
- **Gin** - Web Framework
- **MongoDB** - Database
- **JWT** - Authentication

## API Usage

Main URL / Endpoint :
```
https://your-vercel-url.vercel.app/
```

---

### Authentication

#### Register

```http
POST /api/auth/register
```

```json
{
  "nama": "John Doe",
  "username": "johndoe",
  "password": "123456"
}
```

#### Login

```http
POST /api/auth/login
```

```json
{
  "username": "johndoe",
  "password": "123456"
}
```

---

### User Profile

> **Note:** Semua endpoint di bawah ini membutuhkan header `Authorization: Bearer <token>`

#### Get Profile

```http
GET /api/user/profile
```

#### Update Profile

```http
PUT /api/user/profile
```

```json
{
  "nama": "John Updated",
  "foto": "https://example.com/foto.jpg"
}
```

#### Change Password

```http
PUT /api/user/change-password
```

```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

---

### Transactions

#### Get All Transactions

```http
GET /api/transactions
```

#### Get Transaction by ID

```http
GET /api/transactions/{id}
```

#### Add Transaction (Pemasukan)

```http
POST /api/transactions
```

```json
{
  "tipe": "pemasukan",
  "nominal": 5000000,
  "catatan": "Gaji Januari",
  "tanggal": "2026-01-15"
}
```

#### Add Transaction (Pengeluaran)

```http
POST /api/transactions
```

```json
{
  "tipe": "pengeluaran",
  "nominal": 50000,
  "kategori": "Makanan & Minuman",
  "catatan": "Makan siang",
  "tanggal": "2026-01-18"
}
```

**Kategori yang tersedia:**
- Makanan & Minuman
- Transportasi
- Belanja
- Tagihan
- Hiburan
- Pendidikan
- Kesehatan
- Lainnya

#### Update Transaction

```http
PUT /api/transactions/{id}
```

```json
{
  "nominal": 75000,
  "catatan": "Makan siang + kopi"
}
```

#### Delete Transaction

```http
DELETE /api/transactions/{id}
```

---

### Financial Goals

#### Get All Goals

```http
GET /api/goals
```

#### Get Goal by ID

```http
GET /api/goals/{id}
```

#### Add Goal

```http
POST /api/goals
```

```json
{
  "nama": "Beli Laptop",
  "target_amount": 12000000
}
```

#### Update Goal

```http
PUT /api/goals/{id}
```

```json
{
  "nama": "Beli Laptop Gaming",
  "target_amount": 15000000
}
```

#### Add Progress (Tambah Tabungan)

```http
POST /api/goals/{id}/add
```

```json
{
  "amount": 100000
}
```

#### Delete Goal

```http
DELETE /api/goals/{id}
```

---

### Statistics

#### Get Summary

```http
GET /api/stats/summary
```

Response:
```json
{
  "saldo": 4950000,
  "total_pemasukan": 5000000,
  "total_pengeluaran": 50000
}
```

#### Get Expense by Category

```http
GET /api/stats/expense-by-category
```

#### Get Income vs Expense

```http
GET /api/stats/income-vs-expense
```

---

### Other

#### Health Check

```http
GET /health
```

#### Get Categories

```http
GET /api/categories
```

---

## Author

- Maiys
