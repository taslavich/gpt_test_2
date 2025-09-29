package bidEngine

import (
	"context"
	"sort"

	bidEngineGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/bidEngine"
	pb "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
	utils "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/utils_grpc"
)

func GetWinnerBidInternal_V_2_5(
	ctx context.Context,
	req *bidEngineGrpc.BidEngineRequest_V2_5,
	profitPercent float32,
	globalId string,
	hostname string,
) (*pb.BidResponse, *pb.BidResponse) {
	if len(req.BidResponses) == 0 {
		return &pb.BidResponse{
				Id: req.BidRequest.Id,
				Seatbid: &pb.SeatBid{
					Bid: []*pb.Bid{},
				},
			}, &pb.BidResponse{
				Id: req.BidRequest.Id,
				Seatbid: &pb.SeatBid{
					Bid: []*pb.Bid{},
				},
			}
	}

	impBids := make(map[string][]*pb.Bid)

	for _, bidResponse := range req.BidResponses {
		for _, bid := range bidResponse.Seatbid.Bid {
			impID := bid.GetImpid()
			impBids[impID] = append(impBids[impID], bid)
		}
	}

	if len(impBids) == 0 {
		return &pb.BidResponse{
				Id:      req.BidRequest.Id,
				Seatbid: &pb.SeatBid{Bid: []*pb.Bid{}},
			}, &pb.BidResponse{
				Id: req.BidRequest.Id,
				Seatbid: &pb.SeatBid{
					Bid: []*pb.Bid{},
				},
			}
	}

	seatBid := &pb.SeatBid{}
	seatBidByDspPrice := &pb.SeatBid{}

	for impID, bids := range impBids {
		sort.Slice(bids, func(i, j int) bool {
			return bids[i].GetPrice() > bids[j].GetPrice()
		})

		winningBid := bids[0]

		var bidFloor float32 = 0
		for _, imp := range req.BidRequest.Imp {
			if imp.GetId() == impID {
				bidFloor = imp.GetBidFloor()
				break
			}
		}

		finalPrice, _, err := applyPriceConstraintsAndPercent(
			winningBid.GetPrice(),
			bidFloor,
			profitPercent,
		)
		if err != nil {
			continue
		}

		wrappedNurl := utils.WrapURL(hostname, *winningBid.Nurl, globalId, utils.NURL)
		wrappedBurl := utils.WrapURL(hostname, *winningBid.Burl, globalId, utils.BURL)
		finalBid := &pb.Bid{
			Id:    winningBid.Id,
			Impid: winningBid.Impid,
			Price: &finalPrice,
			Adid:  winningBid.Adid,
			Nurl:  &wrappedNurl,
			Burl:  &wrappedBurl,
		}

		bidByDspPrice := &pb.Bid{
			Id:    winningBid.Id,
			Impid: winningBid.Impid,
			Price: winningBid.Price,
		}

		seatBid.Bid = append(seatBid.Bid, finalBid)
		seatBidByDspPrice.Bid = append(seatBidByDspPrice.Bid, bidByDspPrice)
	}

	bidResponse := &pb.BidResponse{
		Id:      req.BidRequest.Id,
		Seatbid: seatBid,
	}

	bidResponseByDspPrice := &pb.BidResponse{
		Id:      req.BidRequest.Id,
		Seatbid: seatBidByDspPrice,
	}

	return bidResponse, bidResponseByDspPrice
}
