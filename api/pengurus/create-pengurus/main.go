package handler

import (
	"context"
	"time"

	"github.com/foldadjo/PMII_BE/shered/config"
	"github.com/foldadjo/PMII_BE/shered/middleware"
	"github.com/foldadjo/PMII_BE/shered/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var app *fiber.App

func init() {
	config.ConnectDB()

	app = fiber.New()

	api := app.Group("/api")
	auth := api.Group("/pengurus")
	auth.Post("/", Handler)
}

func Handler(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	user := c.Locals("user").(*middleware.Claims)

	// Check if user has permission to create pengurus
	// This is a simple check - you might want to implement more complex permission logic
	if user.Role != "pkn" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
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

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate user exists
	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var users models.User
	err = config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&users)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Validate level
	level := models.PengurusLevel(input.Level)
	switch level {
	case models.LevelPB, models.LevelPKC, models.LevelPC, models.LevelPK, models.LevelPR:
		// Valid level
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pengurus level",
		})
	}

	// Validate required fields based on level
	switch level {
	case models.LevelPKC:
		if input.Wilayah == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Wilayah is required for PKC level",
			})
		}
	case models.LevelPC:
		if input.Cabang == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cabang is required for PC level",
			})
		}
	case models.LevelPK, models.LevelPR:
		if input.Komisariat == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Komisariat is required for PK/PR level",
			})
		}
	}

	// Create pengurus entry
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

	result, err := config.DB.Collection("pengurus").InsertOne(context.Background(), pengurus)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating pengurus entry",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pengurus created successfully",
		"pengurus_id": result.InsertedID,
	})
} 