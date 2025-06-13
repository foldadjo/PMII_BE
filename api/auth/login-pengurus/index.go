package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/foldadjo/PMII_BE/shered/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
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

	db = client.Database("pmii-dev")
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

	// Cari langsung di collection pengurus by email
	var pengurus models.Pengurus
	err := db.Collection("pengurus").FindOne(context.TODO(), bson.M{
		"email": input.Email,
		"aktif": true,
	}).Decode(&pengurus)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verifikasi password
	err = bcrypt.CompareHashAndPassword([]byte(pengurus.PasswordHash), []byte(input.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"pengurus_id": pengurus.ID.Hex(),
		"email":       pengurus.Email,
		"level":       pengurus.Level,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Response sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"pengurus": map[string]interface{}{
			"id":     pengurus.ID.Hex(),
			"email":  pengurus.Email,
			"level":  pengurus.Level,
			"wilayah": pengurus.Wilayah,
			"cabang":  pengurus.Cabang,
			"komisariat": pengurus.Komisariat,
		},
	})
}
