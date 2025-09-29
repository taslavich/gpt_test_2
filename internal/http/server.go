package httpServer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

/*func runServer(cfg *config.BiddingEngineConfig) {

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			"%s:%d",
			cfg.HTTPServer.Host,
			cfg.HTTPServer.Port,
		),
	)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBiddingEngineServiceServer(
		s,
		&biddingEngineWeb.Server{
			ProfitPercent:        cfg.ProfitPercent,
			GetWinnerBidInternal: biddingEngine.GetWinnerBid,
		},
	)

	log.Printf("Server started on %s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// Канал для ошибок
	errChan := make(chan error)

	// Запуск сервера в горутине
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		log.Println("Shutting down gracefully...")
		srv.GracefulStop() // Плавная остановка gRPC
	case err := <-errChan:
		log.Fatalf("Server crashed: %v", err)
	}
}*/

func InitHttpRouter() *chi.Mux {
	httpRouter := chi.NewRouter()
	httpRouter.Use(middleware.Logger)
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Use(middleware.Timeout(60 * time.Second))
	httpRouter.Mount("/debug", middleware.Profiler())

	httpRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {})

	return httpRouter
}

func RunHttpServer(ctx context.Context, router *chi.Mux, host string, port uint16) {
	httpServerAddr := fmt.Sprintf("%s:%d", host, port)
	httpServer := http.Server{
		Addr:    httpServerAddr,
		Handler: router,
	}

	errChan := make(chan error)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		select {
		case <-stop:
			log.Println("Shutting down gracefully...")
			httpServer.Shutdown(ctx)
		case err := <-errChan:
			log.Fatalf("Server crashed: %v", err)
		}
	}()

	log.Printf("Start listening to http://%s/", httpServerAddr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
		log.Fatalf("Can't start server: %v", err)
	}
}

func GetDomain(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host

	return scheme + "://" + host
}
