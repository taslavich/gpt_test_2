package converter_v_2_4

/*
func JSONToGRPC(src *json_ortb_V2_4.BidRequest) *ortb_V2_4.BidRequest {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_4.BidRequest{}

	dst.Id = src.Id
	dst.At = src.At

	if src.Imp != nil {
		grpcImps := make([]*ortb_V2_4.Imp, 0, len(src.Imp))
		for _, jsonImp := range src.Imp {
			if jsonImp != nil {
				grpcImps = append(grpcImps, convertJSONImpToGRPC(jsonImp))
			}
		}
		dst.Imp = grpcImps
	}

	dst.Site = convertJSONSiteToGRPC(src.Site)
	dst.App = convertJSONAppToGRPC(src.App)
	dst.Device = convertJSONDeviceToGRPC(src.Device)

	return dst
}

func convertJSONImpToGRPC(src *json_ortb_V2_4.Imp) *ortb_V2_4.Imp {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_4.Imp{}
	dst.Id = src.Id
	if src.BidFloor != nil {
		dst.BidFloor = src.BidFloor
	}
	dst.Banner = convertJSONBannerToGRPC(src.Banner)
	dst.Native = convertJSONNativeToGRPC(src.Native)
	return dst
}

func convertJSONBannerToGRPC(src *json_ortb_V2_4.Banner) *ortb_V2_4.Banner {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.Banner{}
	dst.W = src.W
	dst.H = src.H
	return dst
}

func convertJSONNativeToGRPC(src *json_ortb_V2_4.Native) *ortb_V2_4.Native {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.Native{}
	dst.Request = src.Request
	return dst
}

func convertJSONSiteToGRPC(src *json_ortb_V2_4.Site) *ortb_V2_4.Site {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.Site{}
	dst.Id = src.Id
	return dst
}

func convertJSONAppToGRPC(src *json_ortb_V2_4.App) *ortb_V2_4.App {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.App{}
	dst.Id = src.Id
	return dst
}

func convertJSONDeviceToGRPC(src *json_ortb_V2_4.Device) *ortb_V2_4.Device {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.Device{}
	dst.Ip = src.Ip

	return dst
}

func GRPCToJSON(src *ortb_V2_4.BidRequest) *json_ortb_V2_4.BidRequest {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_4.BidRequest{}

	dst.Id = src.Id
	dst.At = src.At

	if src.Imp != nil {
		jsonImps := make([]*json_ortb_V2_4.Imp, 0, len(src.Imp))
		for _, grpcImp := range src.Imp {
			if grpcImp != nil {
				jsonImps = append(jsonImps, convertGRPCImpToJSON(grpcImp))
			}
		}
		dst.Imp = jsonImps
	}

	dst.Site = convertGRPCSiteToJSON(src.Site)
	dst.App = convertGRPCAppToJSON(src.App)
	dst.Device = convertGRPCDeviceToJSON(src.Device)

	return dst
}

func convertGRPCImpToJSON(src *ortb_V2_4.Imp) *json_ortb_V2_4.Imp {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_4.Imp{}
	dst.Id = src.Id
	if src.BidFloor != nil {
		dst.BidFloor = src.BidFloor
	}
	dst.Banner = convertGRPCBannerToJSON(src.Banner)
	dst.Native = convertGRPCNativeToJSON(src.Native)
	return dst
}

func convertGRPCBannerToJSON(src *ortb_V2_4.Banner) *json_ortb_V2_4.Banner {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.Banner{}
	dst.W = src.W
	dst.H = src.H
	return dst
}

func convertGRPCNativeToJSON(src *ortb_V2_4.Native) *json_ortb_V2_4.Native {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.Native{}
	dst.Request = src.Request
	return dst
}

func convertGRPCSiteToJSON(src *ortb_V2_4.Site) *json_ortb_V2_4.Site {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.Site{}
	dst.Id = src.Id
	return dst
}

func convertGRPCAppToJSON(src *ortb_V2_4.App) *json_ortb_V2_4.App {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.App{}
	dst.Id = src.Id
	return dst
}

func convertGRPCDeviceToJSON(src *ortb_V2_4.Device) *json_ortb_V2_4.Device {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.Device{}
	dst.Ip = src.Ip
	return dst
}

func JSONBidResponseToGRPC(src *json_ortb_V2_4.BidResponse) *ortb_V2_4.BidResponse {
	if src == nil {
		return nil
	}

	dst := &ortb_V2_4.BidResponse{}
	dst.Id = src.Id
	dst.Seatbid = convertJSONSeatBidToGRPC(src.SeatBid)
	return dst
}

func convertJSONSeatBidToGRPC(src *json_ortb_V2_4.SeatBid) *ortb_V2_4.SeatBid {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.SeatBid{}
	if src.Bid != nil {
		grpcBids := make([]*ortb_V2_4.Bid, 0, len(src.Bid))
		for _, jsonBid := range src.Bid {
			grpcBids = append(grpcBids, convertJSONBidToGRPC(&jsonBid))
		}
		dst.Bid = grpcBids
	}
	return dst
}

func convertJSONBidToGRPC(src *json_ortb_V2_4.Bid) *ortb_V2_4.Bid {
	if src == nil {
		return nil
	}
	dst := &ortb_V2_4.Bid{}
	dst.Id = src.Id
	dst.Impid = src.ImpID
	if src.Price != nil {
		dst.Price = src.Price
	}
	dst.Adid = src.Adid
	return dst
}

func GRPCBidResponseToJSON(src *ortb_V2_4.BidResponse) *json_ortb_V2_4.BidResponse {
	if src == nil {
		return nil
	}

	dst := &json_ortb_V2_4.BidResponse{}
	dst.Id = src.Id
	dst.SeatBid = convertGRPCSeatBidToJSON(src.Seatbid)
	return dst
}

func convertGRPCSeatBidToJSON(src *ortb_V2_4.SeatBid) *json_ortb_V2_4.SeatBid {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.SeatBid{}
	if src.Bid != nil {
		jsonBids := make([]json_ortb_V2_4.Bid, 0, len(src.Bid))
		for _, grpcBid := range src.Bid {
			if grpcBid != nil {
				jsonBids = append(jsonBids, *convertGRPCBidToJSON(grpcBid))
			}
		}
		dst.Bid = jsonBids
	}
	return dst
}

func convertGRPCBidToJSON(src *ortb_V2_4.Bid) *json_ortb_V2_4.Bid {
	if src == nil {
		return nil
	}
	dst := &json_ortb_V2_4.Bid{}
	dst.Id = src.Id
	dst.ImpID = src.Impid
	if src.Price != nil {
		dst.Price = src.Price
	}
	dst.Adid = src.Adid
	return dst
}*/
