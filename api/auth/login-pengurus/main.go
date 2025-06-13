package handler

import (
	"context"
	"os"

	"github.com/foldadjo/PMII_BE/shered/config"
	"github.com/foldadjo/PMII_BE/shered/models"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var app *fiber.App

func init() {
	config.ConnectDB()

	app = fiber.New()

	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/login-pengurus", Handler)
}


func Handler(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Find user
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Find active pengurus entry
	var pengurus models.Pengurus
	err = config.DB.Collection("pengurus").FindOne(context.Background(), bson.M{
		"user_id": user.ID,
		"aktif":   true,
	}).Decode(&pengurus)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User is not an active pengurus",
		})
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id":          user.ID.Hex(),
		"email":            user.Email,
		"role":             user.Role,
		"pengurus_level":   pengurus.Level,
		"pengurus_jabatan": pengurus.Jabatan,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	return c.JSON(fiber.Map{
		"token": tokenString,
		"user": fiber.Map{
			"id":              user.ID,
			"email":           user.Email,
			"full_name":       user.FullName,
			"role":            user.Role,
			"pengurus_level":  pengurus.Level,
			"pengurus_jabatan": pengurus.Jabatan,
		},
	})
}