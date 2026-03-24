// Finance Tracker API - A personal finance tracking API with multi-currency support.
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/sereozha/finance-tracker/docs"
	"github.com/sereozha/finance-tracker/internal/adapters/config"
	"github.com/sereozha/finance-tracker/internal/app"
)

// @title Finance Tracker API
// @version 1.0
// @description A personal finance tracking API with multi-currency support
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@financetracker.local
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create application
	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	// Run application (blocks until shutdown)
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
