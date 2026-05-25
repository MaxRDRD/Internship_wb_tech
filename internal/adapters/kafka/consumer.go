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
	reader *kafkago.Reader
	ingest *usecase.Ingest
}

type eventPayload struct {
	Query      string `json:"query"`
	OccurredAt string `json:"occurred_at"`
}

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

func (c *Consumer) Run(ctx context.Context) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("kafka reader close error: %v", err)
		}
	}()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

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

func decodeEvent(data []byte) (domain.SearchEvent, bool) {
	var payload eventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("event decode error: %v", err)
		return domain.SearchEvent{}, false
	}

	query := strings.TrimSpace(payload.Query)
	if query == "" {
		return domain.SearchEvent{}, false
	}

	occurredAt := time.Now().UTC()
	if payload.OccurredAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, payload.OccurredAt)
		if err == nil {
			occurredAt = parsed.UTC()
		}
	}

	return domain.SearchEvent{Query: query, OccurredAt: occurredAt}, true
}
