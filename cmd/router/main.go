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
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/filter"
	dspRouterGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/dspRouter"
	dspRouterWeb "gitlab.com/twinbid-exchange/RTB-exchange/internal/services/dspRouter/web"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig[config.RouterConfig](ctx)
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

	ruleManager := filter.NewRuleManager()

	fileLoader := filter.NewFileRuleLoader(ruleManager, cfg.DspRulesConfigPath, cfg.SppRulesConfigPath)

	log.Println(cfg.DspRulesConfigPath)
	if err := fileLoader.LoadDSPRules(); err != nil {
		log.Printf("Warning: Failed to load dsp filter rules: %v", err)
	} else {
		log.Println("Filter rules loaded successfully")
	}

	if err := fileLoader.LoadSPPRules(); err != nil {
		log.Printf("Warning: Failed to load spp filter rules: %v", err)
	} else {
		log.Println("Filter rules loaded successfully")
	}

	processor := filter.NewFilterProcessor(ruleManager)

	s := grpc.NewServer()
	dspRouterGrpc.RegisterDspRouterServiceServer(
		s,
		dspRouterWeb.NewServer(
			ruleManager,
			fileLoader,
			processor,
			cfg.DspRulesConfigPath,
			cfg.SppRulesConfigPath,
			cfg.DSPEndpoints_v_2_4,
			redisClient,
			cfg.BidResponsesTimeout,
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
