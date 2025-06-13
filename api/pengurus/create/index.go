package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/foldadjo/PMII_BE/shered/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	adminLevel := models.PengurusLevel(claims.PengurusLevel)

	var input struct {
		UserID       string  `json:"user_id"`
		Level        string  `json:"level"`
		Wilayah      *string `json:"wilayah,omitempty"`
		Cabang       *string `json:"cabang,omitempty"`
		Komisariat   *string `json:"komisariat,omitempty"`
		AlamatSekre  *string `json:"alamat_sekre,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	targetLevel := models.PengurusLevel(input.Level)

	// Cek apakah level target sesuai izin admin
	if !canCreate(adminLevel, targetLevel) {
		http.Error(w, "You don't have permission to create this level", http.StatusForbidden)
		return
	}

	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Cek user exist
	var user models.User
	err = db.Collection("users").FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Validasi data wajib sesuai level
	switch targetLevel {
	case models.LevelPKC:
		if input.Wilayah == nil {
			http.Error(w, "Wilayah required for PKC", http.StatusBadRequest)
			return
		}
	case models.LevelPC:
		if input.Cabang == nil {
			http.Error(w, "Cabang required for PC", http.StatusBadRequest)
			return
		}
	case models.LevelPK, models.LevelPR:
		if input.Komisariat == nil {
			http.Error(w, "Komisariat required for PK/PR", http.StatusBadRequest)
			return
		}
	}

	// Create pengurus baru
	pengurus := models.Pengurus{
		UserID:       userID,
		Level:        targetLevel,
		Wilayah:      input.Wilayah,
		Cabang:       input.Cabang,
		Komisariat:   input.Komisariat,
		AlamatSekre:  input.AlamatSekre,
		Jabatan:      "Ketua",
		Aktif:        true,
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

// Logika izin pembuatan
func canCreate(admin, target models.PengurusLevel) bool {
	switch admin {
	case models.LevelPB:
		return target == models.LevelPKC || target == models.LevelPC || target == models.LevelPK || target == models.LevelPR
	case models.LevelPKC:
		return target == models.LevelPC || target == models.LevelPK || target == models.LevelPR
	case models.LevelPC:
		return target == models.LevelPK || target == models.LevelPR
	case models.LevelPK:
		return target == models.LevelPR
	default:
		return false
	}
}
