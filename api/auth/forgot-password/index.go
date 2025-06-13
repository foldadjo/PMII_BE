package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/foldadjo/PMII_BE/shered/models"
)

// Global mongo client & db
var client *mongo.Client
var db *mongo.Database

func init() {
	mongoURI := os.Getenv("MONGODB_URI")
	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	db = client.Database("pmii-dev") // ganti dengan DB kamu
}

func Handle(w http.ResponseWriter, r *http.Request) {
	// hanya POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.Collection("users").FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Always return generic response to prevent email enumeration
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If your email is registered, you will receive a password reset link",
		})
		return
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Save token
	resetToken := models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err = db.Collection("reset_password_tokens").InsertOne(context.TODO(), resetToken)
	if err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "If your email is registered, you will receive a password reset link",
	})
}
