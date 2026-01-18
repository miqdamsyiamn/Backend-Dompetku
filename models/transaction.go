package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Kategori pengeluaran yang diizinkan
var AllowedCategories = []string{
	"Makanan & Minuman",
	"Transportasi",
	"Belanja",
	"Tagihan",
	"Hiburan",
	"Pendidikan",
	"Kesehatan",
	"Lainnya",
}

type Transaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Tipe      string             `bson:"tipe" json:"tipe"`           // "pemasukan" atau "pengeluaran"
	Nominal   float64            `bson:"nominal" json:"nominal"`
	Kategori  string             `bson:"kategori" json:"kategori"`   // hanya untuk pengeluaran
	Catatan   string             `bson:"catatan" json:"catatan"`
	Tanggal   time.Time          `bson:"tanggal" json:"tanggal"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateTransactionInput struct {
	Tipe     string  `json:"tipe" binding:"required,oneof=pemasukan pengeluaran"`
	Nominal  float64 `json:"nominal" binding:"required,gt=0"`
	Kategori string  `json:"kategori"`
	Catatan  string  `json:"catatan"`
	Tanggal  string  `json:"tanggal" binding:"required"`
}

type UpdateTransactionInput struct {
	Tipe     string  `json:"tipe" binding:"omitempty,oneof=pemasukan pengeluaran"`
	Nominal  float64 `json:"nominal" binding:"omitempty,gt=0"`
	Kategori string  `json:"kategori"`
	Catatan  string  `json:"catatan"`
	Tanggal  string  `json:"tanggal"`
}

// Helper function to check if category is valid
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}
