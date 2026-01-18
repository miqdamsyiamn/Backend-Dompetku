package controllers

import (
	"context"
	"net/http"
	"time"

	"DompetKu/config"
	"DompetKu/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate kategori for pengeluaran
	if input.Tipe == "pengeluaran" {
		if input.Kategori == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kategori wajib diisi untuk pengeluaran"})
			return
		}
		if !models.IsValidCategory(input.Kategori) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":              "Kategori tidak valid",
				"allowed_categories": models.AllowedCategories,
			})
			return
		}
	}

	// Parse tanggal
	tanggal, err := time.Parse("2006-01-02", input.Tanggal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal tidak valid. Gunakan format YYYY-MM-DD"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction := models.Transaction{
		ID:        primitive.NewObjectID(),
		UserID:    objectID,
		Tipe:      input.Tipe,
		Nominal:   input.Nominal,
		Kategori:  input.Kategori,
		Catatan:   input.Catatan,
		Tanggal:   tanggal,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaksi berhasil ditambahkan",
		"transaction": transaction,
	})
}

func GetTransactions(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter by tipe if provided
	filter := bson.M{"user_id": objectID}
	tipe := c.Query("tipe")
	if tipe != "" && (tipe == "pemasukan" || tipe == "pengeluaran") {
		filter["tipe"] = tipe
	}

	// Sort by tanggal descending
	opts := options.Find().SetSort(bson.D{{Key: "tanggal", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

func GetTransactionByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	transactionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

func UpdateTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	transactionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var input models.UpdateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if transaction exists and belongs to user
	var existingTransaction models.Transaction
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID}).Decode(&existingTransaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	// Build update object
	update := bson.M{"updated_at": time.Now()}

	tipe := input.Tipe
	if tipe == "" {
		tipe = existingTransaction.Tipe
	} else {
		update["tipe"] = tipe
	}

	if input.Nominal > 0 {
		update["nominal"] = input.Nominal
	}

	// Validate kategori for pengeluaran
	if tipe == "pengeluaran" {
		kategori := input.Kategori
		if kategori == "" {
			kategori = existingTransaction.Kategori
		}
		if kategori != "" && !models.IsValidCategory(kategori) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":              "Kategori tidak valid",
				"allowed_categories": models.AllowedCategories,
			})
			return
		}
		if input.Kategori != "" {
			update["kategori"] = input.Kategori
		}
	} else if tipe == "pemasukan" {
		update["kategori"] = ""
	}

	if input.Catatan != "" {
		update["catatan"] = input.Catatan
	}

	if input.Tanggal != "" {
		tanggal, err := time.Parse("2006-01-02", input.Tanggal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal tidak valid"})
			return
		}
		update["tanggal"] = tanggal
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Get updated transaction
	var transaction models.Transaction
	collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&transaction)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaksi berhasil diperbarui",
		"transaction": transaction,
	})
}

func DeleteTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	transactionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaksi berhasil dihapus"})
}
