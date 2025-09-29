package converter_v_2_5

/*
func JSONToGRPC(src *json_ortb_V2_5.BidRequest) *ortb_V2_5.BidRequest {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_5.BidRequest{}

	dst.Id = src.Id
	dst.At = src.At

	if src.Imp != nil {
		grpcImps := make([]*ortb_V2_5.Imp, 0, len(src.Imp))
		for _, jsonImp := range src.Imp {
			if jsonImp != nil {
				grpcImps = append(grpcImps, convertJSONImpToGRPC(jsonImp))
			}
		}
		dst.Imp = grpcImps
	}

	dst.Device = convertJSONDeviceToGRPC(src.Device)

	return dst
}

func convertJSONImpToGRPC(src *json_ortb_V2_5.Imp) *ortb_V2_5.Imp {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_5.Imp{}
	dst.Id = src.Id
	if src.BidFloor != nil {
		dst.BidFloor = src.BidFloor
	}
	dst.Banner = convertJSONBannerToGRPC(src.Banner)
	dst.Native = convertJSONNativeToGRPC(src.Native)
	return dst
}

func convertJSONBannerToGRPC(src *json_ortb_V2_5.Banner) *ortb_V2_5.Banner {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_5.Banner{}
	dst.W = src.W
	dst.H = src.H
	return dst
}

func convertJSONNativeToGRPC(src *json_ortb_V2_5.Native) *ortb_V2_5.Native {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_5.Native{}
	dst.Request = src.Request
	return dst
}

func convertJSONDeviceToGRPC(src *json_ortb_V2_5.Device) *ortb_V2_5.Device {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_5.Device{}
	dst.Ip = src.Ip

	return dst
}

func GRPCToJSON(src *ortb_V2_5.BidRequest) *json_ortb_V2_5.BidRequest {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_5.BidRequest{}

	dst.Id = src.Id
	dst.At = src.At

	if src.Imp != nil {
		jsonImps := make([]*json_ortb_V2_5.Imp, 0, len(src.Imp))
		for _, grpcImp := range src.Imp {
			if grpcImp != nil {
				jsonImps = append(jsonImps, convertGRPCImpToJSON(grpcImp))
			}
		}
		dst.Imp = jsonImps
	}

	dst.Device = convertGRPCDeviceToJSON(src.Device)

	return dst
}

func convertGRPCImpToJSON(src *ortb_V2_5.Imp) *json_ortb_V2_5.Imp {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_5.Imp{}
	dst.Id = src.Id
	if src.BidFloor != nil {
		dst.BidFloor = src.BidFloor
	}
	dst.Banner = convertGRPCBannerToJSON(src.Banner)
	dst.Native = convertGRPCNativeToJSON(src.Native)
	return dst
}

func convertGRPCBannerToJSON(src *ortb_V2_5.Banner) *json_ortb_V2_5.Banner {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_5.Banner{}
	dst.W = src.W
	dst.H = src.H
	return dst
}

func convertGRPCNativeToJSON(src *ortb_V2_5.Native) *json_ortb_V2_5.Native {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_5.Native{}
	dst.Request = src.Request
	return dst
}

func convertGRPCDeviceToJSON(src *ortb_V2_5.Device) *json_ortb_V2_5.Device {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_5.Device{}
	dst.Ip = src.Ip
	return dst
}

func JSONBidResponseToGRPC(src *json_ortb_V2_5.BidResponse) *ortb_V2_5.BidResponse {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_5.BidResponse{}
	dst.Id = src.Id
	dst.Seatbid = convertJSONSeatBidToGRPC(src.SeatBid)
	return dst
}

func convertJSONSeatBidToGRPC(src *json_ortb_V2_5.SeatBid) *ortb_V2_5.SeatBid {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_5.SeatBid{}
	if src.Bid != nil {
		grpcBids := make([]*ortb_V2_5.Bid, 0, len(src.Bid))
		for _, jsonBid := range src.Bid {
			grpcBids = append(grpcBids, convertJSONBidToGRPC(&jsonBid))
		}
		dst.Bid = grpcBids
	}
	return dst
}

func convertJSONBidToGRPC(src *json_ortb_V2_5.Bid) *ortb_V2_5.Bid {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_5.Bid{}
	dst.Id = src.Id
	dst.Impid = src.ImpID
	if src.Price != nil {
		dst.Price = src.Price
	}
	dst.Adid = src.Adid
	return dst
}

func GRPCBidResponseToJSON(src *ortb_V2_5.BidResponse) *json_ortb_V2_5.BidResponse {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_5.BidResponse{}
	dst.Id = src.Id
	dst.SeatBid = convertGRPCSeatBidToJSON(src.Seatbid)
	return dst
}

func convertGRPCSeatBidToJSON(src *ortb_V2_5.SeatBid) *json_ortb_V2_5.SeatBid {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_5.SeatBid{}
	if src.Bid != nil {
		jsonBids := make([]json_ortb_V2_5.Bid, 0, len(src.Bid))
		for _, grpcBid := range src.Bid {
			if grpcBid != nil {
				jsonBids = append(jsonBids, *convertGRPCBidToJSON(grpcBid))
			}
		}
		dst.Bid = jsonBids
	}
	return dst
}

func convertGRPCBidToJSON(src *ortb_V2_5.Bid) *json_ortb_V2_5.Bid {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_5.Bid{}
	dst.Id = src.Id
	dst.ImpID = src.Impid
	if src.Price != nil {
		dst.Price = src.Price
	}
	dst.Adid = src.Adid
	return dst
}*/
