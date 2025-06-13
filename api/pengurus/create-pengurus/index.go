package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Middleware parsing token JWT & extract claims
func parseClaimsFromRequest(r *http.Request) (*models.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenStr, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := parseClaimsFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.Role != "pkn" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var input struct {
		UserID       string    `json:"user_id"`
		Level        string    `json:"level"`
		Wilayah      *string   `json:"wilayah,omitempty"`
		Cabang       *string   `json:"cabang,omitempty"`
		Komisariat   *string   `json:"komisariat,omitempty"`
		Jabatan      string    `json:"jabatan"`
		MulaiJabatan time.Time `json:"mulai_jabatan"`
		AkhirJabatan time.Time `json:"akhir_jabatan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var users models.User
	err = db.Collection("users").FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&users)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	level := models.PengurusLevel(input.Level)
	switch level {
	case models.LevelPB, models.LevelPKC, models.LevelPC, models.LevelPK, models.LevelPR:
	default:
		http.Error(w, "Invalid pengurus level", http.StatusBadRequest)
		return
	}

	// Validate required fields per level
	switch level {
	case models.LevelPKC:
		if input.Wilayah == nil {
			http.Error(w, "Wilayah is required for PKC level", http.StatusBadRequest)
			return
		}
	case models.LevelPC:
		if input.Cabang == nil {
			http.Error(w, "Cabang is required for PC level", http.StatusBadRequest)
			return
		}
	case models.LevelPK, models.LevelPR:
		if input.Komisariat == nil {
			http.Error(w, "Komisariat is required for PK/PR level", http.StatusBadRequest)
			return
		}
	}

	pengurus := models.Pengurus{
		UserID:       userID,
		Level:        level,
		Wilayah:      input.Wilayah,
		Cabang:       input.Cabang,
		Komisariat:   input.Komisariat,
		Jabatan:      input.Jabatan,
		Aktif:        true,
		MulaiJabatan: input.MulaiJabatan,
		AkhirJabatan: input.AkhirJabatan,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := db.Collection("pengurus").InsertOne(context.TODO(), pengurus)
	if err != nil {
		http.Error(w, "Error creating pengurus", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Pengurus created successfully",
		"pengurus_id": result.InsertedID,
	})
}
