package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/foldadjo/PMII_BE/shered/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client
var db *mongo.Database

func init() {
	mongoURI := os.Getenv("MONGODB_URI")
	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	db = client.Database("pmii-dev") // Ganti dengan nama database
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

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id":           user.ID.Hex(),
		"email":             user.Email,
		"role":              user.Role,
		"code_kepengurusan": user.CodeKepengurusan,
		"exp":               time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Clean user response (jangan return password hash)
	userResponse := map[string]interface{}{
		"id":                user.ID.Hex(),
		"email":             user.Email,
		"full_name":         user.FullName,
		"anggota_id":        user.AnggotaID,
		"role":              user.Role,
		"code_kepengurusan": user.CodeKepengurusan,
		"gender":            user.Gender,
		"birth_day":         user.BirthDay,
		"active":            user.Active,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user":  userResponse,
	})
}
