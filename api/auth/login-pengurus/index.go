package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.Collection("users").FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Cari pengurus aktif
	var pengurus models.Pengurus
	err = db.Collection("pengurus").FindOne(context.TODO(), bson.M{
		"user_id": user.ID,
		"aktif":   true,
	}).Decode(&pengurus)
	if err != nil {
		http.Error(w, "User is not an active pengurus", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id":          user.ID.Hex(),
		"email":            user.Email,
		"role":             user.Role,
		"pengurus_level":   pengurus.Level,
		"pengurus_jabatan": pengurus.Jabatan,
		"exp":              time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":              user.ID.Hex(),
			"email":           user.Email,
			"full_name":       user.FullName,
			"role":            user.Role,
			"pengurus_level":  pengurus.Level,
			"pengurus_jabatan": pengurus.Jabatan,
		},
	})
}
