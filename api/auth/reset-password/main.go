package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/awslabs/aws-lambda-go-api-proxy/fiberadapter"
	"github.com/foldadjo/PMII_BE/shered/config"
	"github.com/foldadjo/PMII_BE/shered/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var adapter *fiberadapter.FiberLambda

func init() {
	config.ConnectDB()

	app := fiber.New()

	app.Post("/api/auth/reset-password", ReserPassword)

	// Buat adapter untuk Vercel
	adapter = fiberadapter.New(app)
}

func ReserPassword(c *fiber.Ctx) error {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Find valid reset token
	var resetToken models.ResetPasswordToken
	err := config.DB.Collection("reset_password_tokens").FindOne(context.Background(), bson.M{
		"token":      input.Token,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&resetToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	// Update user password
	_, err = config.DB.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": resetToken.UserID},
		bson.M{"$set": bson.M{
			"password_hash": string(hashedPassword),
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error updating password",
		})
	}

	// Delete used token
	_, err = config.DB.Collection("reset_password_tokens").DeleteOne(context.Background(), bson.M{"_id": resetToken.ID})
	if err != nil {
		// Log error but don't return it to user
		log.Printf("Error deleting reset token: %v", err)
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successful",
	})
} 

// Exported handler untuk Vercel
func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	adapter.ProxyWithContext(ctx, w, r)
}
