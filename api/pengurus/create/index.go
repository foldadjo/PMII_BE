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

	// request body struct
	var input struct {
		Email        string  `json:"email"`
		Password     string  `json:"password"`
		NoTelp       *string `json:"no_telp,omitempty"`
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

	// Check role permission
	if !canCreate(adminLevel, targetLevel) {
		http.Error(w, "You don't have permission to create this level", http.StatusForbidden)
		return
	}

	// Check existing email
	var existing models.Pengurus
	err = db.Collection("pengurus").FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&existing)
	if err == nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Validate mandatory fields based on level
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

	// Insert Pengurus
	pengurus := models.Pengurus{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		NoTelp:       input.NoTelp,
		Level:        targetLevel,
		Wilayah:      input.Wilayah,
		Cabang:       input.Cabang,
		Komisariat:   input.Komisariat,
		AlamatSekre:  input.AlamatSekre,
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

// Role permission logic
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
