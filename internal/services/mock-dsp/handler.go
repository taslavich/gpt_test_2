package mockDspWeb

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ggicci/httpin"
	"github.com/unrolled/render"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
)

var rnr = render.New(render.Options{
	StreamingJSON: true,
})

func getHealth(
	w http.ResponseWriter,
) {
	w.WriteHeader(http.StatusOK)
}

func postBid_V2_5(
	ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
	resp *ortb_V2_5.BidResponse,
) {
	startTime := time.Now()
	input := r.Context().Value(httpin.Input).(*postBidRequest_V2_5)

	var statusCode int

	if len(input.Payload.Imp) > 0 {
		resp.Seatbid.Bid[0].Impid = input.Payload.Imp[0].Id
		statusCode = http.StatusOK
	} else {
		resp.Seatbid.Bid = []*ortb_V2_5.Bid{}
		statusCode = http.StatusNoContent
	}

	if err := rnr.JSON(w, statusCode, postBidResponse_V2_5{
		BidResponse: resp,
	}); err != nil {
		log.Printf("[%s] Cannot make HTTP response back: %v\n", resp.GetId(), err)
	}

	processingTime := time.Since(startTime)
	log.Printf("[%s] Request processed in %v", resp.GetId(), processingTime)
}
