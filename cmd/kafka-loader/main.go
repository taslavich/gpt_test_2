package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/config"
	kafka_loader "gitlab.com/twinbid-exchange/RTB-exchange/internal/services/kafka-loader"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig[config.KafkaLoaderConfig](ctx)
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}
	log.Println("Config initialized!")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBroker),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		Async:        false,
		BatchTimeout: 100 * time.Millisecond,
	}
	defer kafkaWriter.Close()

	if err := kafka_loader.EnsureTopicExists(cfg.KafkaBroker, cfg.KafkaTopic); err != nil {
		log.Fatalf("⚠️ Failed to ensure topic exists: %v", err)
	} else {
		log.Printf("✅ Kafka topic %s is ready", cfg.KafkaTopic)
	}

	log.Println("✅ Kafka writer initialized")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var totalProcessed int64
	ticker := time.NewTicker(time.Duration(cfg.FlushIntervalSec) * time.Second)
	defer ticker.Stop()

	log.Printf("🚀 Kafka Loader started. Processing every %d seconds", cfg.FlushIntervalSec)

	for {
		select {
		case <-sigChan:
			log.Printf("🛑 Shutting down Kafka Loader. Total processed: %d records", totalProcessed)
			return
		case <-ticker.C:
			err := kafka_loader.ProcessBatch(context.Background(), redisClient, kafkaWriter, cfg.BatchSize)
			if err != nil {
				log.Printf("❌ Batch processing error: %v", err)
				continue
			}
		}
	}
}
