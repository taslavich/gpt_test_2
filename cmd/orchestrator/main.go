package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/config"
	orchestratorGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/orchestrator"
	orchestrator "gitlab.com/twinbid-exchange/RTB-exchange/internal/services/orchestrator/service"
	orchestratorWeb "gitlab.com/twinbid-exchange/RTB-exchange/internal/services/orchestrator/web"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig[config.OrchestratorConfig](ctx)
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

	o := orchestrator.NewOrchestrator(cfg.UriOfBidEngine, cfg.UriOfDspRouter)

	clients, cancelFunc := o.GetGrpClients()
	log.Println(clients.DspRouterGrpcClient.GetBids_V2_5(ctx, nil))

	defer cancelFunc()

	s := grpc.NewServer()
	orchestratorGrpc.RegisterOrchestratorServiceServer(
		s,
		orchestratorWeb.NewServer(
			clients.BidEngineGrpcClient,
			clients.DspRouterGrpcClient,
			redisClient,
			cfg.GetBidsTimeout,
			cfg.AuctionTimeout,
		),
	)

	errChan := make(chan error)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		select {
		case <-stop:
			log.Println("Shutting down gracefully...")
			s.GracefulStop()
		case err := <-errChan:
			log.Fatalf("Server crashed: %v", err)
		}
	}()

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			"%s:%d",
			cfg.Host,
			cfg.Port,
		),
	)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server started on %s:%d", cfg.Host, cfg.Port)
	if err := s.Serve(lis); err != nil {
		errChan <- err
		log.Printf("failed to serve: %v", err)
	}
}
