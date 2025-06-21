
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port string
	
	// Operation configuration (needed for handling operations)
	DefaultCount   int
	DefaultTimeout time.Duration
	MaxCount       int
	MaxTimeout     time.Duration
	EnableLogging  bool
	
	// PocketBase configuration
	PocketBaseEnabled bool
	PocketBaseURL     string
	
	// Regional Agent configuration
	RegionName      string
	AgentID         string
	AgentIPAddress  string
	Token           string
	
	// Monitoring configuration
	CheckInterval   time.Duration
	MaxRetries      int
	RequestTimeout  time.Duration
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
		log.Printf("Using system environment variables instead")
	} else {
		//log.Printf("Successfully loaded .env file")
	}

	return &Config{
		Port:              getEnv("PORT", "8091"),
		DefaultCount:      getIntEnv("DEFAULT_COUNT", 4),
		DefaultTimeout:    getDurationEnv("DEFAULT_TIMEOUT", 10*time.Second),
		MaxCount:          getIntEnv("MAX_COUNT", 20),
		MaxTimeout:        getDurationEnv("MAX_TIMEOUT", 30*time.Second),
		EnableLogging:     getBoolEnv("ENABLE_LOGGING", true),
		PocketBaseEnabled: getBoolEnv("POCKETBASE_ENABLED", true),
		PocketBaseURL:     getEnv("POCKETBASE_URL", "http://localhost:8090"),
		RegionName:        getEnv("REGION_NAME", ""),        // No default - must be set
		AgentID:           getEnv("AGENT_ID", ""),           // No default - must be set
		AgentIPAddress:    getEnv("AGENT_IP_ADDRESS", ""),   // No default - must be set
		Token:             getEnv("AGENT_TOKEN", ""),
		CheckInterval:     getDurationEnv("CHECK_INTERVAL", 30*time.Second),
		MaxRetries:        getIntEnv("MAX_RETRIES", 3),
		RequestTimeout:    getDurationEnv("REQUEST_TIMEOUT", 10*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}