package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/db"
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/operation"
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/rabbitmq"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/metrics"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// Options for CLI
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8082"`
}

var dbConn *gorm.DB

func main() {
	_ = godotenv.Load()
	dbConn = db.Init()
	// RabbitMQ setup
	var conn *amqp.Connection
	var ch *amqp.Channel
	disableRabbit := os.Getenv("DISABLE_RABBITMQ") == "true"

	if !disableRabbit {
		conn, ch = rabbitmq.Connect()
		eventRouter := rabbitmq.SetupEventHandlers(dbConn)
		go func() {
			if _, err := rabbitmq.StartListening(ch, eventRouter); err != nil {
				log.Fatalf("Failed to start event listener: %v", err)
			}
		}()
	} else {
		log.Println("DISABLE_RABBITMQ=true, skipping RabbitMQ connection")
	}

	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()

		router.Use(middleware.Logger)
		router.Use(middleware.Recoverer)
		router.Use(middleware.Compress(5))

		router.Use(metrics.Collector(metrics.CollectorOpts{
			Host:  false,
			Proto: true,
			Skip: func(r *http.Request) bool {
				return r.Method != "OPTIONS"
			},
		}))

		router.Handle("/metrics", metrics.Handler())

		configs := huma.DefaultConfig("Paye Ton Kawa - Orders", "1.0.0")
		api := humachi.New(router, configs)

		operation.RegisterOrdersRoutes(api, dbConn, ch)

		// Create the HTTP server.
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: router,
		}

		// OnStart: blocking ListenAndServe
		hooks.OnStart(func() {
			log.Printf("Starting server on port %d...", options.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server failed: %v", err)
			}
		})

		// OnStop: graceful shutdown
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Printf("Server shutdown error: %v", err)
			}
			if conn != nil {
				_ = conn.Close()
			}
			if ch != nil {
				_ = ch.Close()
			}
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
