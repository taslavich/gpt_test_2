package bidEngineWeb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/constants"
	bidEngineGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/bidEngine"
	pb "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/bidEngine"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
	utils "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/utils_grpc"
)

type Server struct {
	ProfitPercent float32
	redisClient   *redis.Client
	timeout       time.Duration
	hostname      string

	GetWinnerBidInternal_V_2_4 func(
		ctx context.Context,
		req *bidEngineGrpc.BidEngineRequest_V2_4,
		profitPercent float32,
		globalId string,
		hostname string,
	) (*ortb_V2_4.BidResponse, *ortb_V2_4.BidResponse)

	GetWinnerBidInternal_V_2_5 func(
		ctx context.Context,
		req *bidEngineGrpc.BidEngineRequest_V2_5,
		profitPercent float32,
		globalId string,
		hostname string,
	) (*ortb_V2_5.BidResponse, *ortb_V2_5.BidResponse)

	pb.BidEngineServiceServer
}

func NewServer(
	ProfitPercent float32,
	redisClient *redis.Client,
	hostname string,
	GetWinnerBidInternal_V_2_4 func(
		ctx context.Context,
		req *bidEngineGrpc.BidEngineRequest_V2_4,
		profitPercent float32,
		globalId string,
		hostname string,
	) (*ortb_V2_4.BidResponse, *ortb_V2_4.BidResponse),
	GetWinnerBidInternal_V_2_5 func(
		ctx context.Context,
		req *bidEngineGrpc.BidEngineRequest_V2_5,
		profitPercent float32,
		globalId string,
		hostname string,
	) (*ortb_V2_5.BidResponse, *ortb_V2_5.BidResponse),
) *Server {
	return &Server{
		ProfitPercent:              ProfitPercent,
		redisClient:                redisClient,
		hostname:                   hostname,
		GetWinnerBidInternal_V_2_4: GetWinnerBidInternal_V_2_4,
		GetWinnerBidInternal_V_2_5: GetWinnerBidInternal_V_2_5,
	}
}

func (s *Server) GetWinnerBid_V2_4(
	ctx context.Context,
	req *bidEngineGrpc.BidEngineRequest_V2_4,
) (
	*bidEngineGrpc.BidEngineResponse_V2_4,
	error,
) {
	bidResponse, bidResponseByDspPrice := s.GetWinnerBidInternal_V_2_4(
		ctx,
		req,
		s.ProfitPercent,
		req.GlobalId,
		s.hostname,
	)

	data, err := json.Marshal(bidResponse)
	if err != nil {
		fmt.Printf("failed to marshal JSON in GetWinnerBidInternal: %w", err)
	}

	if err := utils.WriteJsonToRedis(ctx, s.redisClient, req.GlobalId, constants.BID_RESPONSE_WINNER_COLUMN, data); err != nil {
		fmt.Printf("failed to WriteJsonToRedis Bid BID_RESPONSE_WINNER in GetWinnerBidInternal: %w", err)
	}

	dataByDspPrice, err := json.Marshal(bidResponseByDspPrice)
	if err != nil {
		fmt.Printf("failed to marshal JSON in GetWinnerBidInternal: %w", err)
	}

	if err := utils.WriteJsonToRedis(ctx, s.redisClient, req.GlobalId, constants.BID_RESPONSE_WINNER_BY_DSP_PRICE_COLUMN, dataByDspPrice); err != nil {
		fmt.Printf("failed to WriteJsonToRedis Bid BID_RESPONSE_WINNER_BY_DSP_PRICE_COLUMN in GetWinnerBidInternal: %w", err)
	}

	return &bidEngineGrpc.BidEngineResponse_V2_4{
		BidResponse: bidResponse,
	}, nil
}
