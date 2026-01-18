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

func CreateGoal(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.CreateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	goal := models.FinancialGoal{
		ID:            primitive.NewObjectID(),
		UserID:        objectID,
		Nama:          input.Nama,
		TargetAmount:  input.TargetAmount,
		CurrentAmount: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = collection.InsertOne(ctx, goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create goal"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Goal berhasil dibuat",
		"goal":    goal,
		"progress_percentage": goal.GetProgressPercentage(),
	})
}

func GetGoals(c *gin.Context) {
	userID, _ := c.Get("userID")
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{"user_id": objectID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch goals"})
		return
	}
	defer cursor.Close(ctx)

	var goals []models.FinancialGoal
	if err := cursor.All(ctx, &goals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode goals"})
		return
	}

	// Add progress percentage to each goal
	type GoalWithProgress struct {
		models.FinancialGoal
		ProgressPercentage float64 `json:"progress_percentage"`
	}

	var goalsWithProgress []GoalWithProgress
	for _, g := range goals {
		goalsWithProgress = append(goalsWithProgress, GoalWithProgress{
			FinancialGoal:      g,
			ProgressPercentage: g.GetProgressPercentage(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"goals": goalsWithProgress,
		"count": len(goals),
	})
}

func GetGoalByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	goalID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var goal models.FinancialGoal
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID}).Decode(&goal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goal":                goal,
		"progress_percentage": goal.GetProgressPercentage(),
	})
}

func UpdateGoal(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	goalID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var input models.UpdateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if goal exists and belongs to user
	var existingGoal models.FinancialGoal
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID}).Decode(&existingGoal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal tidak ditemukan"})
		return
	}

	update := bson.M{"updated_at": time.Now()}
	if input.Nama != "" {
		update["nama"] = input.Nama
	}
	if input.TargetAmount > 0 {
		update["target_amount"] = input.TargetAmount
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update goal"})
		return
	}

	// Get updated goal
	var goal models.FinancialGoal
	collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&goal)

	c.JSON(http.StatusOK, gin.H{
		"message":             "Goal berhasil diperbarui",
		"goal":                goal,
		"progress_percentage": goal.GetProgressPercentage(),
	})
}

func AddProgressToGoal(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	goalID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var input models.AddProgressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if goal exists and belongs to user
	var goal models.FinancialGoal
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID}).Decode(&goal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal tidak ditemukan"})
		return
	}

	newAmount := goal.CurrentAmount + input.Amount

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{
			"current_amount": newAmount,
			"updated_at":     time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add progress"})
		return
	}

	goal.CurrentAmount = newAmount

	c.JSON(http.StatusOK, gin.H{
		"message":             "Tabungan berhasil ditambahkan",
		"goal":                goal,
		"progress_percentage": goal.GetProgressPercentage(),
	})
}

func DeleteGoal(c *gin.Context) {
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	goalID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	collection := config.GetCollection("financial_goals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID, "user_id": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete goal"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal berhasil dihapus"})
}
