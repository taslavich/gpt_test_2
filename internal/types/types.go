package types

type StatisticsRecord struct {
	BID_REQUEST                      string `json:"BID_REQUEST"`
	GEO_COLUMN                       string `json:"GEO_COLUMN"`
	BID_RESPONSES                    string `json:"BID_RESPONSES"`
	BID_RESPONSE_WINNER              string `json:"BID_RESPONSE_WINNER"`
	BID_RESPONSE_WINNER_BY_DSP_PRICE string `json:"BID_RESPONSE_WINNER_BY_DSP_PRICE"`
	SUCCESS                          string `json:"SUCCESS"`
}
