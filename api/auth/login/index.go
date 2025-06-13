package handler

import (
	"context"
	"net/http"
	"os"

	"github.com/awslabs/aws-lambda-go-api-proxy/fiberadapter"
	"github.com/foldadjo/PMII_BE/shered/config"
	"github.com/foldadjo/PMII_BE/shered/models"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var adapter *fiberadapter.FiberLambda

func init() {
	config.ConnectDB()

	app := fiber.New()

	app.Post("/api/auth/login", Login)

	// Buat adapter untuk Vercel
	adapter = fiberadapter.New(app)
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

// Exported handler untuk Vercel
func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	adapter.ProxyWithContext(ctx, w, r)
}