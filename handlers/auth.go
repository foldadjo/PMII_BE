package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jauharmaknun/PMII_BE/config"
	"github.com/jauharmaknun/PMII_BE/models"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if email exists
	var existingUser models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	// Create user
	user := models.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FullName:     input.FullName,
		Role:         models.RoleNull,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := config.DB.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user_id": result.InsertedID,
	})
}

func Login(c *fiber.Ctx) error {
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

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"role":    user.Role,
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
			"id":       user.ID,
			"email":    user.Email,
			"full_name": user.FullName,
			"role":     user.Role,
		},
	})
}

func LoginPengurus(c *fiber.Ctx) error {
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

func ForgotPassword(c *fiber.Ctx) error {
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
		// Return success even if user not found to prevent enumeration
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

	// In a real application, send email here

	return c.JSON(fiber.Map{
		"message": "If your email is registered, you will receive a password reset link",
	})
}

func ResetPassword(c *fiber.Ctx) error {
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