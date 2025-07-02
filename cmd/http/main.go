package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"mateo/internal/domain"
	"mateo/internal/service/invoice"
	"mateo/internal/service/merchant"
	"mateo/internal/service/requisite"
	"mateo/internal/store/pg"
	"mateo/internal/store/pgcached"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"mateo/internal/config"
	"mateo/internal/transport/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Создаем пул подключений вместо одиночного соединения
	pgConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse DB config")
	}

	// Настройки пула (рекомендуемые значения)
	pgConfig.MaxConns = 50
	pgConfig.MinConns = 10
	pgConfig.MaxConnLifetime = time.Hour
	pgConfig.MaxConnIdleTime = time.Minute * 30
	pgConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create connection pool")
	}
	defer pool.Close() // Важно: закрываем пул при завершении

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	// Initialize store с использованием пула
	store := pg.NewStore(pool)

	// Init Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close() // Закрываем подключение к Redis

	cachedStore := pgcached.NewCachedStore(store, redisClient)

	// Initialize services
	merchantService := merchant.NewService(cachedStore)
	invoiceService := invoice.NewService(cachedStore)
	requisiteService := requisite.NewService(cachedStore)

	// Initialize app
	app := domain.NewApp(merchantService, requisiteService, invoiceService)

	// Initialize and start HTTP server
	srv, err := http.NewServer(app)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Start server in a goroutine
	go func() {
		if err := srv.Start(cfg.HTTP.Port); err != nil {
			log.Error().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	if err := srv.Stop(cfg.HTTP.ShutdownTimeout); err != nil {
		log.Error().Err(err).Msg("Failed to stop server")
	}

	log.Info().Msg("Server stopped")
}
