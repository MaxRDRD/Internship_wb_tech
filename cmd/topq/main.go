package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"topq/internal/adapters/http"
	"topq/internal/adapters/kafka"
	"topq/internal/adapters/memory"
	"topq/internal/config"
	"topq/internal/usecase"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем файл .env в переменные окружения (если он присутствует)
	// чтобы не экспортировать переменные вручную в среде выполнения
	_ = godotenv.Load()

	// загружаем конфиг
	cfg := config.Load()

	//инициализируем repo и usecase
	repo := memory.NewSlidingWindowRepo(cfg.WindowSeconds)
	stopRepo := memory.NewStopListRepo()
	stopList := usecase.NewStopList(stopRepo)
	antiSpam := usecase.NewAntiSpam(10, time.Minute)
	ingest := usecase.NewIngest(repo, stopRepo, antiSpam)
	top := usecase.NewTop(repo, stopRepo)

	// запускаем http сервер и kafka consumer
	server := http.NewServer(cfg.HTTPAddr, top, stopList, cfg.DefaultTopN, cfg.WindowSeconds)
	// осуществляем graceful shutdown при получении сигнала завершения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.ConsumerEnabled {
		consumer := kafka.NewConsumer(cfg, ingest)
		go func() {
			if err := consumer.Run(ctx); err != nil {
				log.Printf("kafka consumer stopped: %v", err)
			}
		}()
	} else {
		log.Printf("kafka consumer disabled: set KAFKA_BROKERS and KAFKA_TOPIC")
	}

	if err := server.Run(ctx); err != nil {
		log.Printf("http server stopped: %v", err)
	}

	<-ctx.Done()
	time.Sleep(50 * time.Millisecond)
}
