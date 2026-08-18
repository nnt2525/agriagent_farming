package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"agriagent/backend/internal/agent"
	"agriagent/backend/internal/config"
	"agriagent/backend/internal/handlers"
	"agriagent/backend/internal/llm"
	"agriagent/backend/internal/mqtt"
	"agriagent/backend/internal/store"
)

func main() {
	_ = godotenv.Load() // .env is optional; real env vars always win

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st := buildStore(ctx, cfg)
	llmClient := buildLLMClient(cfg)

	// ag.ExecuteFunc is left nil for now: relay actuation would publish an
	// MQTT command topic once the ESP32 relay firmware defines one. The
	// agent still logs and records every auto-executed decision either way.
	ag := agent.New(cfg, st, llmClient)
	go ag.Run(ctx)

	if cfg.MQTTBrokerURL != "" {
		sub := mqtt.New(cfg, st, ag)
		if err := sub.Connect(); err != nil {
			log.Printf("mqtt: failed to connect (continuing without live telemetry): %v", err)
		} else {
			defer sub.Disconnect()
		}
	} else {
		log.Println("mqtt: MQTT_BROKER_URL not set, skipping live telemetry (mock/demo mode)")
	}

	app := fiber.New(fiber.Config{
		AppName: "AgriAgent API",
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET,POST,OPTIONS",
	}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "mock_mode": cfg.MockMode})
	})

	api := app.Group("/api")
	handlers.New(st, ag).Register(api)

	log.Printf("AgriAgent backend listening on :%s (mock_mode=%v)", cfg.Port, cfg.MockMode)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func buildStore(ctx context.Context, cfg config.Config) store.Store {
	if !cfg.MockMode && cfg.DatabaseURL != "" {
		pg, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to connect to postgres (set MOCK_MODE=true to run without a DB): %v", err)
		}
		log.Println("store: using PostgresStore")
		return pg
	}
	log.Println("store: using in-memory MockStore (set MOCK_MODE=false and DATABASE_URL to use Postgres)")
	return store.NewMockStore()
}

func buildLLMClient(cfg config.Config) llm.Client {
	if cfg.GeminiAPIKey == "" {
		log.Println("llm: GEMINI_API_KEY not set, using offline MockClient")
		return llm.NewMockClient()
	}
	log.Printf("llm: using Gemini client (model=%s)", cfg.GeminiModel)
	return llm.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
}
