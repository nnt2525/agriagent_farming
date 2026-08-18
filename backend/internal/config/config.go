package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all environment-driven settings for the service.
type Config struct {
	Port string

	// MockMode: when true, handlers serve in-memory seeded data instead of
	// querying Postgres. Lets the frontend integrate before Supabase/MQTT
	// are wired up. Set MOCK_MODE=false once DATABASE_URL is real.
	MockMode bool

	DatabaseURL string

	MQTTBrokerURL string // e.g. tls://xxxx.s1.eu.hivemq.cloud:8883
	MQTTUsername  string
	MQTTPassword  string
	MQTTClientID  string
	MQTTTopic     string // subscribe pattern, e.g. agriagent/+/telemetry

	GeminiAPIKey string
	GeminiModel  string

	// Agent thresholds
	SoilMoistureLowPct   float64       // trigger watering below this
	SoilMoistureDeltaPct float64       // trigger on sudden change greater than this
	ScheduleInterval     time.Duration // periodic re-evaluation (default 3h)
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		MockMode:    getEnvBool("MOCK_MODE", true),
		DatabaseURL: getEnv("DATABASE_URL", ""),

		MQTTBrokerURL: getEnv("MQTT_BROKER_URL", ""),
		MQTTUsername:  getEnv("MQTT_USERNAME", ""),
		MQTTPassword:  getEnv("MQTT_PASSWORD", ""),
		MQTTClientID:  getEnv("MQTT_CLIENT_ID", "agriagent-backend"),
		MQTTTopic:     getEnv("MQTT_TOPIC", "agriagent/+/telemetry"),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-2.0-flash"),

		SoilMoistureLowPct:   getEnvFloat("SOIL_MOISTURE_LOW_PCT", 30.0),
		SoilMoistureDeltaPct: getEnvFloat("SOIL_MOISTURE_DELTA_PCT", 10.0),
		ScheduleInterval:     getEnvDuration("AGENT_SCHEDULE_INTERVAL", 3*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
