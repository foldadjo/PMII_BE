package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/foldadjo/PMII_BE/config"
	"github.com/foldadjo/PMII_BE/handlers"
	"github.com/foldadjo/PMII_BE/middleware"

	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var app *fiber.App

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config.ConnectDB()

	app = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/login-pengurus", handlers.LoginPengurus)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/reset-password", handlers.ResetPassword)

	protected := api.Group("/", middleware.Protected())
	protected.Post("/create-pengurus", handlers.CreatePengurus)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fasthttpadaptor.NewFastHTTPHandler(app.Handler())(w, r)
}
