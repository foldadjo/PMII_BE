package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/awslabs/aws-lambda-go-api-proxy/fiberadapter"
	"github.com/foldadjo/PMII_BE/shered/config"
	"github.com/foldadjo/PMII_BE/shered/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var adapter *fiberadapter.FiberLambda

func init() {
	// Connect DB sekali saat cold start
	config.ConnectDB()

	// Setup Fiber
	app := fiber.New()

	// Daftar route
	app.Post("/api/auth/forgot-password", ForgotPasswordHandler)

	// Buat adapter untuk Vercel
	adapter = fiberadapter.New(app)
}

func ForgotPasswordHandler(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
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
		return c.JSON(fiber.Map{
			"message": "If your email is registered, you will receive a password reset link",
		})
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating reset token",
		})
	}
	token := hex.EncodeToString(tokenBytes)

	// Save reset token
	resetToken := models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err = config.DB.Collection("reset_password_tokens").InsertOne(context.Background(), resetToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error saving reset token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "If your email is registered, you will receive a password reset link",
	})
}

// Exported handler untuk Vercel
func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	adapter.ProxyWithContext(ctx, w, r)
}
