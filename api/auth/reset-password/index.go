package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/foldadjo/PMII_BE/shered/models"
)

// Global MongoDB client
var client *mongo.Client
var db *mongo.Database

func init() {
	mongoURI := os.Getenv("MONGODB_URI")
	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	db = client.Database("your_db_name") // ganti sesuai database kamu
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Cari reset token valid
	var resetToken models.ResetPasswordToken
	err := db.Collection("reset_password_tokens").FindOne(context.TODO(), bson.M{
		"token":      input.Token,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&resetToken)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Update user password
	_, err = db.Collection("users").UpdateOne(
		context.TODO(),
		bson.M{"_id": resetToken.UserID},
		bson.M{"$set": bson.M{
			"password_hash": string(hashedPassword),
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Hapus token setelah digunakan
	_, err = db.Collection("reset_password_tokens").DeleteOne(context.TODO(), bson.M{"_id": resetToken.ID})
	if err != nil {
		log.Printf("Error deleting reset token: %v", err)
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successful",
	})
}
