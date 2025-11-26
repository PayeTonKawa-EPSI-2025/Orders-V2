package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	"gorm.io/gorm"
)

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8082"`
}

var (
	dbConn *gorm.DB
)

func main() {
	_ = godotenv.Load()
	dbConn = db.Init()
	conn, ch := rabbitmq.Connect()

	// Set up event handlers
	eventRouter := rabbitmq.SetupEventHandlers(dbConn)

	// Start listening for events
	_, err := rabbitmq.StartListening(ch, eventRouter)
	if err != nil {
		log.Fatalf("Failed to start event listener: %v", err)
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
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: router,
		}

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			server.ListenAndServe()
		})

		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)

			// Close the RabbitMQ connection when server shuts down
			conn.Close()
			ch.Close()
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
