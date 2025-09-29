package main

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/twinbid-exchange/RTB-exchange/internal/config"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
	httpServer "gitlab.com/twinbid-exchange/RTB-exchange/internal/http"
	mockDspWeb "gitlab.com/twinbid-exchange/RTB-exchange/internal/services/mock-dsp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig[config.MockDspConfig](ctx)
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}
	log.Println("Config initialized!")

	log.Println(cfg.DspName)

	BidId := fmt.Sprint(cfg.DspName, cfg.DspName)
	Nurl := "Nurl"
	Burl := "Burl"
	router := httpServer.InitHttpRouter()
	mockDspWeb.InitRoutes(
		ctx,
		router,
		&ortb_V2_5.BidResponse{
			Id: &cfg.DspName,
			Seatbid: &ortb_V2_5.SeatBid{
				Bid: []*ortb_V2_5.Bid{
					{
						Id:    &BidId,
						Price: &cfg.Price,
						Adid:  &cfg.Adid,
						Nurl:  &Nurl,
						Burl:  &Burl,
					},
				},
			},
		},
	)
	log.Println("HTTP routes initialized")

	httpServer.RunHttpServer(ctx, router, cfg.Host, cfg.Port)

}
