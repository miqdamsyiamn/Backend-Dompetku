package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FinancialGoal struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Nama          string             `bson:"nama" json:"nama"`
	TargetAmount  float64            `bson:"target_amount" json:"target_amount"`
	CurrentAmount float64            `bson:"current_amount" json:"current_amount"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateGoalInput struct {
	Nama         string  `json:"nama" binding:"required"`
	TargetAmount float64 `json:"target_amount" binding:"required,gt=0"`
}

type UpdateGoalInput struct {
	Nama         string  `json:"nama" binding:"omitempty"`
	TargetAmount float64 `json:"target_amount" binding:"omitempty,gt=0"`
}

type AddProgressInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawProgressInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// Helper function to calculate progress percentage
func (g *FinancialGoal) GetProgressPercentage() float64 {
	if g.TargetAmount == 0 {
		return 0
	}
	percentage := (g.CurrentAmount / g.TargetAmount) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}
