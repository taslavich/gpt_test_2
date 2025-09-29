package sppAdapterWeb

import (
	"context"
	"net/http"
	"time"

	orchestratorProto "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/orchestrator"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"

	"github.com/ggicci/httpin"
	"github.com/ggicci/httpin/integration"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

const (
	PostBid_V_2_4_URL = "/bid_v_2_4"
	PostBid_V_2_5_URL = "/bid_v_2_5"

	GetNurlUrl = "/nurl"
	GetBurlUrl = "/burl"

	GetHealthUrl = "/health"
)

type postBidRequest_V2_4 struct {
	Payload *ortb_V2_4.BidRequest `in:"body=json"`
}

type postBidResponse_V2_4 struct {
	*ortb_V2_4.BidResponse
}

type postBidRequest_V2_5 struct {
	Payload *ortb_V2_5.BidRequest `in:"body=json"`
}

type postBidResponse_V2_5 struct {
	*ortb_V2_5.BidResponse
}

type nurlRequest struct {
	GlobalId string `in:"query=id"`
	DspURL   string `in:"query=url"`
}

type burlRequest struct {
	GlobalId string `in:"query=id"`
	DspURL   string `in:"query=url"`
}

func InitRoutes(
	ctx context.Context,
	httpRouter *chi.Mux,
	redisClient *redis.Client,
	isBadIp func(ipStr string) (bool, error),
	getCountryISO func(ipStr string) (string, error),
	orchestratorClient orchestratorProto.OrchestratorServiceClient,
	bidRequestTimeout,
	nurlTimeout,
	burlTimeout time.Duration,
) {
	integration.UseGochiURLParam("path", chi.URLParam)

	httpRouter.With(
		httpin.NewInput(postBidRequest_V2_4{}),
	).Post(PostBid_V_2_4_URL, func(w http.ResponseWriter, r *http.Request) {
		postBid_V2_4(ctx, w, r, redisClient, isBadIp, getCountryISO, orchestratorClient, bidRequestTimeout)
	})

	httpRouter.With(
		httpin.NewInput(postBidRequest_V2_5{}),
	).Post(PostBid_V_2_5_URL, func(w http.ResponseWriter, r *http.Request) {
		postBid_V2_5(ctx, w, r, redisClient, isBadIp, getCountryISO, orchestratorClient, bidRequestTimeout)
	})

	httpRouter.Get(GetNurlUrl, func(w http.ResponseWriter, r *http.Request) {
		getNurl(ctx, w, r, nurlTimeout)
	})

	httpRouter.Get(GetBurlUrl, func(w http.ResponseWriter, r *http.Request) {
		getBurl(ctx, w, r, redisClient, burlTimeout)
	})

	httpRouter.Get(GetHealthUrl, func(w http.ResponseWriter, r *http.Request) {
		getHealth(w)
	})
}
