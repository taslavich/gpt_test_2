package json_ortb_V2_5

type BidRequest struct {
	Id     *string `json:"id"`
	At     *int32  `json:"at"`
	Imp    []*Imp  `json:"imp"`
	Device *Device `json:"device,omitempty"`
}

type Imp struct {
	Id       *string  `json:"id"`
	BidFloor *float32 `json:"bidFloor"`
	Banner   *Banner  `json:"banner"`
	Native   *Native  `json:"native"`
}

type Banner struct {
	W *int32 `json:"w"`
	H *int32 `json:"h"`
}

type Native struct {
	Request *string `json:"request"`
}

type Device struct {
	Ip  *string `json:"ip"`
	Geo *Geo    `json:"geo"`
}

type Geo struct {
	Country *string `json:"country"`
}

type BidResponse struct {
	Id      *string  `json:"id"`
	SeatBid *SeatBid `json:"seatbid,omitempty"`
}

type SeatBid struct {
	Bid []Bid `json:"bid"`
}

type Bid struct {
	Id    *string  `json:"id"`
	ImpID *string  `json:"impid"`
	Price *float32 `json:"price"`
	Adid  *string  `json:"adid,omitempty"`
	Nurl  *string  `json:"nurl,omitempty"`
	Burl  *string  `json:"burl,omitempty"`
}
