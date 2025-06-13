package handler

import (
	"context"
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

	app.Post("/api/auth/register", Register)

	// Buat adapter untuk Vercel
	adapter = fiberadapter.New(app)
}

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

func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	adapter.ProxyWithContext(ctx, w, r)
}