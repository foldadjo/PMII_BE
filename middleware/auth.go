package middleware

import (
	"strings"
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	PengurusLevel string `json:"pengurus_level,omitempty"`
	PengurusJabatan string `json:"pengurus_jabatan,omitempty"`
	jwt.RegisteredClaims
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			c.Locals("user", claims)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}
} 