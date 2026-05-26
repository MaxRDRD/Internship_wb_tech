package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr     = ":8080" // стандартный порт для kafka ui
	defaultWindowSec    = 300     // стаднартный размер скользящего окна
	defaultTopN         = 10      //стандартное количество интерсующих запросов в топе
	defaultKafkaGroupID = "topq"  // стандартный ID группы kafka
)

type Config struct {
	HTTPAddr        string // порт
	WindowSeconds   int    // размер скользящего окна
	DefaultTopN     int    //количество интерсующих запросов в топе
	ConsumerEnabled bool   // Флаг, указывающий, нужно ли запускать Kafka Consumer

	KafkaBrokers []string // Список брокеров Kafka
	KafkaTopic   string   // Топик Kafka для чтения данных
	KafkaGroupID string   // ID группы Kafka Consumer
}

// Загрузка конфига
func Load() Config {
	cfg := Config{}
	// Читаем конфигурацию из переменных окружения с дефолтными значениями
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

// хелпер функции для получение данных из переменных окружения
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
