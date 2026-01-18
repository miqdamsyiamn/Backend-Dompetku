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
)

func GetSummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate untuk menghitung total pemasukan dan pengeluaran
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": objectID}},
		{"$group": bson.M{
			"_id":   "$tipe",
			"total": bson.M{"$sum": "$nominal"},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode summary"})
		return
	}

	var totalPemasukan, totalPengeluaran float64
	for _, r := range results {
		if r["_id"] == "pemasukan" {
			totalPemasukan = r["total"].(float64)
		} else if r["_id"] == "pengeluaran" {
			totalPengeluaran = r["total"].(float64)
		}
	}

	saldo := totalPemasukan - totalPengeluaran

	c.JSON(http.StatusOK, gin.H{
		"saldo":            saldo,
		"total_pemasukan":  totalPemasukan,
		"total_pengeluaran": totalPengeluaran,
	})
}

func GetExpenseByCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate pengeluaran per kategori
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id": objectID,
			"tipe":    "pengeluaran",
		}},
		{"$group": bson.M{
			"_id":   "$kategori",
			"total": bson.M{"$sum": "$nominal"},
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"total": -1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get expense by category"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
		return
	}

	// Format untuk pie chart
	type CategoryData struct {
		Kategori string  `json:"kategori"`
		Total    float64 `json:"total"`
		Count    int32   `json:"count"`
	}

	var categories []CategoryData
	var grandTotal float64

	for _, r := range results {
		kategori := ""
		if r["_id"] != nil {
			kategori = r["_id"].(string)
		}
		total := r["total"].(float64)
		count := r["count"].(int32)

		categories = append(categories, CategoryData{
			Kategori: kategori,
			Total:    total,
			Count:    count,
		})
		grandTotal += total
	}

	// Tambahkan persentase
	type CategoryWithPercentage struct {
		Kategori   string  `json:"kategori"`
		Total      float64 `json:"total"`
		Count      int32   `json:"count"`
		Percentage float64 `json:"percentage"`
	}

	var categoriesWithPercentage []CategoryWithPercentage
	for _, cat := range categories {
		percentage := 0.0
		if grandTotal > 0 {
			percentage = (cat.Total / grandTotal) * 100
		}
		categoriesWithPercentage = append(categoriesWithPercentage, CategoryWithPercentage{
			Kategori:   cat.Kategori,
			Total:      cat.Total,
			Count:      cat.Count,
			Percentage: percentage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"categories":  categoriesWithPercentage,
		"grand_total": grandTotal,
	})
}

func GetIncomeVsExpense(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.GetCollection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate pemasukan vs pengeluaran
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": objectID}},
		{"$group": bson.M{
			"_id":   "$tipe",
			"total": bson.M{"$sum": "$nominal"},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get income vs expense"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
		return
	}

	var pemasukan, pengeluaran float64
	var countPemasukan, countPengeluaran int32

	for _, r := range results {
		if r["_id"] == "pemasukan" {
			pemasukan = r["total"].(float64)
			countPemasukan = r["count"].(int32)
		} else if r["_id"] == "pengeluaran" {
			pengeluaran = r["total"].(float64)
			countPengeluaran = r["count"].(int32)
		}
	}

	grandTotal := pemasukan + pengeluaran
	pemasukanPercentage := 0.0
	pengeluaranPercentage := 0.0

	if grandTotal > 0 {
		pemasukanPercentage = (pemasukan / grandTotal) * 100
		pengeluaranPercentage = (pengeluaran / grandTotal) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"tipe":       "pemasukan",
				"total":      pemasukan,
				"count":      countPemasukan,
				"percentage": pemasukanPercentage,
			},
			{
				"tipe":       "pengeluaran",
				"total":      pengeluaran,
				"count":      countPengeluaran,
				"percentage": pengeluaranPercentage,
			},
		},
		"grand_total": grandTotal,
	})
}

// GetCategories returns all allowed expense categories
func GetCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories": models.AllowedCategories,
	})
}
