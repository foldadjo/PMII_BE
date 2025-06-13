package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Domain struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Domain    string            `bson:"domain" json:"domain"`
	IsActive  bool              `bson:"is_active" json:"is_active"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
} 