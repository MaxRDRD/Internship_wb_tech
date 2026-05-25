package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr     = ":8080"
	defaultWindowSec    = 300
	defaultTopN         = 10
	defaultKafkaGroupID = "topq"
)

type Config struct {
	HTTPAddr        string
	WindowSeconds   int
	DefaultTopN     int
	ConsumerEnabled bool

	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
}

func Load() Config {
	cfg := Config{}

	cfg.HTTPAddr = getenvDefault("HTTP_ADDR", defaultHTTPAddr)
	cfg.WindowSeconds = getenvIntDefault("WINDOW_SECONDS", defaultWindowSec)
	cfg.DefaultTopN = getenvIntDefault("DEFAULT_TOP_N", defaultTopN)
	cfg.KafkaGroupID = getenvDefault("KAFKA_GROUP", defaultKafkaGroupID)

	brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if brokers != "" {
		parts := strings.Split(brokers, ",")
		for _, broker := range parts {
			broker = strings.TrimSpace(broker)
			if broker != "" {
				cfg.KafkaBrokers = append(cfg.KafkaBrokers, broker)
			}
		}
	}
	cfg.KafkaTopic = strings.TrimSpace(os.Getenv("KAFKA_TOPIC"))

	cfg.ConsumerEnabled = len(cfg.KafkaBrokers) > 0 && cfg.KafkaTopic != ""
	return cfg
}

func getenvDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvIntDefault(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
