package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/yourusername/device-inventory/config"
	"github.com/yourusername/device-inventory/internal/handlers"
	"github.com/yourusername/device-inventory/internal/repository"
	"github.com/yourusername/device-inventory/internal/service"
)

func main() {
	// Load .env file if it exists (won't error if missing — good for production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 1. Connect to the database
	db := config.ConnectDB()
	defer db.Close()

	// 2. Wire up dependencies manually (dependency injection without a framework)
	//    This pattern is idiomatic Go — no magic, no reflection, just functions.
	//
	//    Flow: DB → Repository → Service → Handler
	//    Each layer only knows about the layer directly below it.
	deviceRepo := repository.NewDeviceRepository(db)
	deviceService := service.NewDeviceService(deviceRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceService)

	// 3. Create the Fiber app
	app := fiber.New(fiber.Config{
		// Custom error handler so all errors return consistent JSON
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 4. Global middleware
	app.Use(logger.New())   // Logs every request: method, path, status, latency
	app.Use(recover.New())  // Catches panics so the server doesn't crash

	// 5. Health check endpoint — simple but expected in any real service
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "device-inventory",
		})
	})

	// 6. Register device routes
	deviceHandler.RegisterRoutes(app)

	// 7. Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
