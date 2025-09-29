package filter

import (
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
)

// V24BidRequestExtractor экстрактор для ORTB v2.4
type V24BidRequestExtractor struct {
	req *ortb_V2_4.BidRequest
}

func NewV24BidRequestExtractor(req *ortb_V2_4.BidRequest) *V24BidRequestExtractor {
	return &V24BidRequestExtractor{req: req}
}

func (e *V24BidRequestExtractor) ExtractFieldValue(field FieldType) FieldValue {
	switch field {
	case FieldBidFloor:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].BidFloor != nil {
					return NewFloatValue(float64(*e.req.Imp[i].BidFloor))
				}
			}
		}
		return NewFloatValue(0)

	case FieldAppID:
		if e.req.App != nil && e.req.App.Id != nil {
			return NewStringValue(*e.req.App.Id)
		}
		return NewStringValue("")

	case FieldSiteID:
		if e.req.Site != nil && e.req.Site.Id != nil {
			return NewStringValue(*e.req.Site.Id)
		}
		return NewStringValue("")

	case FieldDeviceIP:
		if e.req.Device != nil && e.req.Device.Ip != nil {
			return NewStringValue(*e.req.Device.Ip)
		}
		return NewStringValue("")

	case FieldBannerWidth:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].Banner != nil && e.req.Imp[i].Banner.W != nil {
					return NewIntValue(int(*e.req.Imp[i].Banner.W))
				}
			}
		}
		return NewIntValue(0)

	case FieldBannerHeight:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].Banner != nil && e.req.Imp[i].Banner.H != nil {
					return NewIntValue(int(*e.req.Imp[i].Banner.H))
				}
			}
		}
		return NewIntValue(0)

	case FieldDeviceCountry:
		if e.req.Device != nil && e.req.Device.Geo != nil && e.req.Device.Geo.Country != nil {
			return NewStringValue(*e.req.Device.Geo.Country)
		}
		return NewStringValue("")

	default:
		return FieldValue{}
	}
}

// V25BidRequestExtractor экстрактор для ORTB v2.5
type V25BidRequestExtractor struct {
	req *ortb_V2_5.BidRequest
}

func NewV25BidRequestExtractor(req *ortb_V2_5.BidRequest) *V25BidRequestExtractor {
	return &V25BidRequestExtractor{req: req}
}

func (e *V25BidRequestExtractor) ExtractFieldValue(field FieldType) FieldValue {
	switch field {
	case FieldBidFloor:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].BidFloor != nil {
					return NewFloatValue(float64(*e.req.Imp[i].BidFloor))
				}
			}
		}
		return NewFloatValue(0)

	case FieldAppID:
		// V2.5 не имеет App, возвращаем пустое значение
		return NewStringValue("")

	case FieldSiteID:
		// V2.5 не имеет Site, возвращаем пустое значение
		return NewStringValue("")

	case FieldDeviceIP:
		if e.req.Device != nil && e.req.Device.Ip != nil {
			return NewStringValue(*e.req.Device.Ip)
		}
		return NewStringValue("")

	case FieldBannerWidth:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].Banner != nil && e.req.Imp[i].Banner.W != nil {
					return NewIntValue(int(*e.req.Imp[i].Banner.W))
				}
			}
		}
		return NewIntValue(0)

	case FieldBannerHeight:
		if e.req.Imp != nil {
			for i := range e.req.Imp {
				if e.req.Imp[i] != nil && e.req.Imp[i].Banner != nil && e.req.Imp[i].Banner.H != nil {
					return NewIntValue(int(*e.req.Imp[i].Banner.H))
				}
			}
		}
		return NewIntValue(0)

	case FieldDeviceCountry:
		if e.req.Device != nil && e.req.Device.Geo != nil && e.req.Device.Geo.Country != nil {
			return NewStringValue(*e.req.Device.Geo.Country)
		}
		return NewStringValue("")

	default:
		return FieldValue{}
	}
}

// V24BidResponseExtractor экстрактор для BidResponse v2.4
type V24BidResponseExtractor struct {
	resp *ortb_V2_4.BidResponse
}

func NewV24BidResponseExtractor(resp *ortb_V2_4.BidResponse) *V24BidResponseExtractor {
	return &V24BidResponseExtractor{resp: resp}
}

func (e *V24BidResponseExtractor) ExtractFieldValue(field FieldType) FieldValue {
	switch field {
	case FieldBidPrice:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Price != nil {
					return NewFloatValue(float64(*e.resp.Seatbid.Bid[i].Price))
				}
			}
		}
		return NewFloatValue(0)

	case FieldBidID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Id != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Id)
				}
			}
		}
		return NewStringValue("")

	case FieldBidAdID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Adid != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Adid)
				}
			}
		}
		return NewStringValue("")

	case FieldBidImpID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Impid != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Impid)
				}
			}
		}
		return NewStringValue("")

	case FieldBidArray:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			return NewStringValue("exists")
		}
		return NewStringValue("")

	case FieldBidNurl:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Nurl != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Nurl)
				}
			}
		}
		return NewStringValue("")

	case FieldBidBurl:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Burl != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Burl)
				}
			}
		}
		return NewStringValue("")

	default:
		return FieldValue{}
	}
}

// V25BidResponseExtractor экстрактор для BidResponse v2.5
type V25BidResponseExtractor struct {
	resp *ortb_V2_5.BidResponse
}

func NewV25BidResponseExtractor(resp *ortb_V2_5.BidResponse) *V25BidResponseExtractor {
	return &V25BidResponseExtractor{resp: resp}
}

func (e *V25BidResponseExtractor) ExtractFieldValue(field FieldType) FieldValue {
	switch field {
	case FieldBidPrice:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Price != nil {
					return NewFloatValue(float64(*e.resp.Seatbid.Bid[i].Price))
				}
			}
		}
		return NewFloatValue(0)

	case FieldBidID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Id != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Id)
				}
			}
		}
		return NewStringValue("")

	case FieldBidAdID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Adid != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Adid)
				}
			}
		}
		return NewStringValue("")

	case FieldBidImpID:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Impid != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Impid)
				}
			}
		}
		return NewStringValue("")

	case FieldBidArray:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			return NewStringValue("exists")
		}
		return NewStringValue("")

	case FieldBidNurl:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Nurl != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Nurl)
				}
			}
		}
		return NewStringValue("")

	case FieldBidBurl:
		if e.resp != nil && e.resp.Seatbid != nil && e.resp.Seatbid.Bid != nil {
			for i := range e.resp.Seatbid.Bid {
				if e.resp.Seatbid.Bid[i].Burl != nil {
					return NewStringValue(*e.resp.Seatbid.Bid[i].Burl)
				}
			}
		}
		return NewStringValue("")

	default:
		return FieldValue{}
	}
}
