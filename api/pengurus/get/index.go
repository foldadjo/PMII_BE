package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/foldadjo/PMII_BE/shered/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global MongoDB client (reuse dari sebelumnya)
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameter level
	levelParam := r.URL.Query().Get("level")
	if levelParam == "" {
		http.Error(w, "Missing 'level' query parameter", http.StatusBadRequest)
		return
	}

	// Validasi level input
	level := models.PengurusLevel(strings.ToUpper(levelParam))
	switch level {
	case models.LevelPB, models.LevelPKC, models.LevelPC, models.LevelPK, models.LevelPR:
	default:
		http.Error(w, "Invalid level", http.StatusBadRequest)
		return
	}

	// Query database
	filter := bson.M{"level": level, "aktif": true}
	cursor, err := db.Collection("pengurus").Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var results []models.Pengurus
	if err := cursor.All(context.TODO(), &results); err != nil {
		http.Error(w, "Error decoding data", http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":   len(results),
		"pengurus": results,
	})
}
