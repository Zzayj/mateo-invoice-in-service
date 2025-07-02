package http

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"log"
	"time"

	"mateo/internal/domain"

	_ "github.com/lib/pq"
)

// Server represents the HTTP server
type Server struct {
	fiber *fiber.App
	app   *domain.App
}

// NewServer creates a new HTTP server
func NewServer(app *domain.App) (*Server, error) {
	f := fiber.New()
	s := &Server{
		fiber: f,
		app:   app,
	}

	// Middleware
	f.Use(recover.New())
	f.Use(logger.New())

	// API v1 group
	api := f.Group("/api")

	api.Post("/invoice-in", s.CreateInvoice)

	return s, nil
}

// Start starts the HTTP server
func (s *Server) Start(port string) error {
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	return s.fiber.Listen(addr)
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(timeout time.Duration) error {
	return s.fiber.ShutdownWithTimeout(timeout)
}
