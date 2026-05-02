package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server
	GatewayAddr string
	APIAddr     string

	// Database
	DatabaseURL            string
	DatabaseMaxConnections int

	// Redis
	RedisURL string

	// Kafka
	KafkaBrokers       string
	KafkaTopicSignals  string
	KafkaTopicJobs     string
	KafkaConsumerGroup string

	// Qdrant
	QdrantURL        string
	QdrantCollection string

	// Anthropic
	AnthropicAPIKey string
	AnthropicModel  string

	// Auth
	JWTSecret          string
	SessionSecret      string
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	SessionMaxAge      time.Duration

	// Encryption
	EncryptionKey string

	// ML Service
	MLServiceURL string

	// Feature flags
	FeatureSimilarityEnabled   bool
	FeatureAutoCorrelation     bool
	FeatureComplianceReports   bool

	// Limits
	FreeTierIncidentsPerMonth int
	FreeTierLLMCallsPerMonth  int
	IngestionRateLimitPerMin  int

	// Frontend
	FrontendURL string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		GatewayAddr: envOrDefault("GATEWAY_ADDR", ":8080"),
		APIAddr:     envOrDefault("API_ADDR", ":8081"),

		DatabaseURL:            envOrDefault("DATABASE_URL", "postgres://signalroot:signalroot@localhost:5432/signalroot?sslmode=disable"),
		DatabaseMaxConnections: envIntOrDefault("DATABASE_MAX_CONNECTIONS", 25),

		RedisURL: envOrDefault("REDIS_URL", "redis://localhost:6379"),

		KafkaBrokers:       envOrDefault("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicSignals:  envOrDefault("KAFKA_TOPIC_SIGNALS", "signalroot.signals"),
		KafkaTopicJobs:     envOrDefault("KAFKA_TOPIC_JOBS", "signalroot.jobs"),
		KafkaConsumerGroup: envOrDefault("KAFKA_CONSUMER_GROUP", "signalroot-worker"),

		QdrantURL:        envOrDefault("QDRANT_URL", "http://localhost:6333"),
		QdrantCollection: envOrDefault("QDRANT_COLLECTION", "incident_dna"),

		AnthropicAPIKey: envOrDefault("ANTHROPIC_API_KEY", ""),
		AnthropicModel:  envOrDefault("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),

		JWTSecret:          envOrDefault("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		SessionSecret:      envOrDefault("SESSION_SECRET", "dev-session-secret-change-in-production"),
		GoogleClientID:     envOrDefault("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: envOrDefault("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     envOrDefault("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOrDefault("GITHUB_CLIENT_SECRET", ""),
		SessionMaxAge:      time.Duration(envIntOrDefault("SESSION_MAX_AGE_HOURS", 24)) * time.Hour,

		EncryptionKey: envOrDefault("ENCRYPTION_KEY", "dev-encryption-key-32bytes!!!!!"),

		MLServiceURL: envOrDefault("ML_SERVICE_URL", "http://localhost:8082"),

		FeatureSimilarityEnabled: envBoolOrDefault("FEATURE_SIMILARITY_ENABLED", true),
		FeatureAutoCorrelation:   envBoolOrDefault("FEATURE_AUTO_CORRELATION", true),
		FeatureComplianceReports: envBoolOrDefault("FEATURE_COMPLIANCE_REPORTS", true),

		FreeTierIncidentsPerMonth: envIntOrDefault("FREE_TIER_INCIDENTS_PER_MONTH", 50),
		FreeTierLLMCallsPerMonth:  envIntOrDefault("FREE_TIER_LLM_CALLS_PER_MONTH", 100),
		IngestionRateLimitPerMin:  envIntOrDefault("INGESTION_RATE_LIMIT_PER_MINUTE", 1000),

		FrontendURL: envOrDefault("FRONTEND_URL", "http://localhost:3000"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return strings.EqualFold(v, "true") || v == "1"
	}
	return fallback
}
