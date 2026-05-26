package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"topq/internal/config"
	"topq/internal/domain"
	"topq/internal/usecase"
)

type Consumer struct {
	reader *kafkago.Reader // Kafka reader для чтения сообщений из топика
	ingest *usecase.Ingest // usecase для обработки событий и их сохранения в репозитории
}

type eventPayload struct {
	Query      string `json:"query"`       // поисковый запрос
	SessionID  string `json:"session_id"`  // идентификатор сессии, может быть пустым
	UserID     string `json:"user_id"`     // идентификатор пользователя, может быть пустым
	OccurredAt string `json:"occurred_at"` // время события в формате RFC3339Nano, может быть пустым (тогда будет использовано время обработки)
}

// Создание нового Kafka consumer
func NewConsumer(cfg config.Config, ingest *usecase.Ingest) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     cfg.KafkaBrokers,
		Topic:       cfg.KafkaTopic,
		GroupID:     cfg.KafkaGroupID,
		StartOffset: kafkago.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})

	return &Consumer{reader: reader, ingest: ingest}
}

// Запуск Kafka consumer
func (c *Consumer) Run(ctx context.Context) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("kafka reader close error: %v", err)
		}
	}()

	// Бесконечный цикл для чтения сообщений из Kafka
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		// Декодируем сообщение в структуру события
		event, ok := decodeEvent(msg.Value)
		if !ok {
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := c.ingest.Handle(ctx, event); err != nil {
			log.Printf("event ingest error: %v", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("kafka commit error: %v", err)
		}
	}
}

// Декодирование данных из Kafka в структуру SearchEvent
func decodeEvent(data []byte) (domain.SearchEvent, bool) {
	var payload eventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("event decode error: %v", err)
		return domain.SearchEvent{}, false
	}
	// получение запроса из payload
	query := strings.TrimSpace(payload.Query)
	if query == "" {
		return domain.SearchEvent{}, false
	}
	// получение sessionID и userID из payload
	sessionID := strings.TrimSpace(payload.SessionID)
	userID := strings.TrimSpace(payload.UserID)
	// получение времени события из payload
	occurredAt := time.Now().UTC()
	if payload.OccurredAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, payload.OccurredAt)
		if err == nil {
			occurredAt = parsed.UTC()
		}
	}

	return domain.SearchEvent{Query: query, SessionID: sessionID, UserID: userID, OccurredAt: occurredAt}, true
}
