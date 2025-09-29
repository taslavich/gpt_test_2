package mockDspWeb

import (
	"context"
	"net/http"

	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"

	"github.com/ggicci/httpin"
	"github.com/ggicci/httpin/integration"
	"github.com/go-chi/chi/v5"
)

const (
	PostBidURL   = "/bid"
	GetHealthUrl = "/health"
)

type postBidRequest_V2_5 struct {
	Payload *ortb_V2_5.BidRequest `in:"body=json"`
}

type postBidResponse_V2_5 struct {
	*ortb_V2_5.BidResponse
}

func InitRoutes(
	ctx context.Context,
	httpRouter *chi.Mux,
	resp *ortb_V2_5.BidResponse,
) {
	integration.UseGochiURLParam("path", chi.URLParam)

	httpRouter.With(
		httpin.NewInput(postBidRequest_V2_5{}),
	).Post(PostBidURL, func(w http.ResponseWriter, r *http.Request) {
		postBid_V2_5(ctx, r, w, resp)
	})

	httpRouter.Get(GetHealthUrl, func(w http.ResponseWriter, r *http.Request) {
		getHealth(w)
	})
}
