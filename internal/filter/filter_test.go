package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
)

type FilterTestSuite struct {
	suite.Suite
	ruleManager *RuleManager
	processor   *FilterProcessor
	loader      *FileRuleLoader
}

func TestFilterSuite(t *testing.T) {
	suite.Run(t, new(FilterTestSuite))
}

func (suite *FilterTestSuite) SetupTest() {
	suite.ruleManager = NewRuleManager()
	suite.loader = NewFileRuleLoader(suite.ruleManager, "test_dsp.json", "test_spp.json")
	suite.processor = NewFilterProcessor(suite.ruleManager)

	// Загружаем правила перед каждым тестом
	err := suite.loader.LoadDSPRules()
	assert.NoError(suite.T(), err, "Failed to load DSP rules")

	err = suite.loader.LoadSPPRules()
	assert.NoError(suite.T(), err, "Failed to load SPP rules")
}

func (suite *FilterTestSuite) TestDSPFilteringV24() {
	t := suite.T()

	tests := []struct {
		name            string
		dspID           string
		country         string
		bidFloor        float32
		appID           string
		bannerWidth     int32
		deviceIP        string
		expectedAllowed bool
	}{
		{
			name:            "DSP1 should allow US country with bidfloor > 0.5 and app exists",
			dspID:           "dsp1",
			country:         "US",
			bidFloor:        0.6,
			appID:           "app123",
			bannerWidth:     728,
			deviceIP:        "", // DSP1 не требует IP
			expectedAllowed: true,
		},
		{
			name:            "DSP1 should reject non-US country",
			dspID:           "dsp1",
			country:         "CA",
			bidFloor:        0.6,
			appID:           "app123",
			bannerWidth:     728,
			deviceIP:        "", // DSP1 не требует IP
			expectedAllowed: false,
		},
	}

	suite.SetupTest()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := suite.createTestBidRequestV24(tt.country, tt.bidFloor, tt.appID, tt.bannerWidth, 90, tt.deviceIP)
			result := suite.processor.ProcessRequestForDSPV24(tt.dspID, req)

			assert.Equal(t, tt.expectedAllowed, result.Allowed,
				"Test case: %s, Country: %s, BidFloor: %.2f, AppID: %s, BannerWidth: %d, DeviceIP: %s",
				tt.name, tt.country, tt.bidFloor, tt.appID, tt.bannerWidth, tt.deviceIP)
		})
	}
}

func (suite *FilterTestSuite) TestDSPFilteringV25() {
	t := suite.T()

	tests := []struct {
		name            string
		dspID           string
		country         string
		bidFloor        float32
		bannerWidth     int32
		deviceIP        string
		expectedAllowed bool
	}{
		{
			name:            "DSP2 should allow banner width between 300-800 for V2.5",
			dspID:           "dsp2",
			country:         "US",
			bidFloor:        0.3,
			bannerWidth:     500,
			deviceIP:        "192.168.1.1",
			expectedAllowed: true,
		},
		{
			name:            "DSP2 should reject banner width outside range for V2.5",
			dspID:           "dsp2",
			country:         "US",
			bidFloor:        0.3,
			bannerWidth:     200,
			deviceIP:        "192.168.1.1",
			expectedAllowed: false,
		},
	}

	suite.SetupTest()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := suite.createTestBidRequestV25(tt.country, tt.bidFloor, tt.bannerWidth, 90, tt.deviceIP)
			result := suite.processor.ProcessRequestForDSPV25(tt.dspID, req)

			assert.Equal(t, tt.expectedAllowed, result.Allowed,
				"Test case: %s, Country: %s, BidFloor: %.2f, BannerWidth: %d, DeviceIP: %s",
				tt.name, tt.country, tt.bidFloor, tt.bannerWidth, tt.deviceIP)
		})
	}
}

func (suite *FilterTestSuite) TestSPPFilteringV24() {
	t := suite.T()

	tests := []struct {
		name            string
		sppID           string
		bidPrice        float32
		bidID           string
		adID            string
		impID           string
		hasBids         bool
		expectedAllowed bool
	}{
		{
			name:            "SPP1 should allow bid price > 1.0 with bids for V2.4",
			sppID:           "spp1",
			bidPrice:        1.5,
			bidID:           "bid123",
			adID:            "ad456",
			impID:           "imp789",
			hasBids:         true,
			expectedAllowed: true,
		},
	}

	suite.SetupTest()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := suite.createTestBidResponseV24(tt.bidPrice, tt.bidID, tt.adID, tt.impID, tt.hasBids)
			result := suite.processor.ProcessResponseForSPPV24(tt.sppID, resp)

			assert.Equal(t, tt.expectedAllowed, result.Allowed,
				"Test case: %s, BidPrice: %.2f, AdID: %s, HasBids: %v",
				tt.name, tt.bidPrice, tt.adID, tt.hasBids)
		})
	}
}

func (suite *FilterTestSuite) TestSPPFilteringV25() {
	t := suite.T()

	tests := []struct {
		name            string
		sppID           string
		bidPrice        float32
		bidID           string
		adID            string
		impID           string
		hasBids         bool
		expectedAllowed bool
	}{
		{
			name:            "SPP2 should allow when adid not equal to test-ad for V2.5",
			sppID:           "spp2",
			bidPrice:        1.0,
			bidID:           "bid123",
			adID:            "different-ad",
			impID:           "imp789",
			hasBids:         true,
			expectedAllowed: true,
		},
	}

	suite.SetupTest()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := suite.createTestBidResponseV25(tt.bidPrice, tt.bidID, tt.adID, tt.impID, tt.hasBids)
			result := suite.processor.ProcessResponseForSPPV25(tt.sppID, resp)

			assert.Equal(t, tt.expectedAllowed, result.Allowed,
				"Test case: %s, BidPrice: %.2f, AdID: %s, HasBids: %v",
				tt.name, tt.bidPrice, tt.adID, tt.hasBids)
		})
	}
}

// Вспомогательные методы для создания тестовых данных
func (suite *FilterTestSuite) createTestBidRequestV24(country string, bidFloor float32, appID string, bannerWidth, bannerHeight int32, deviceIP string) *ortb_V2_4.BidRequest {
	return &ortb_V2_4.BidRequest{
		Device: &ortb_V2_4.Device{
			Ip: &deviceIP,
			Geo: &ortb_V2_4.Geo{
				Country: &country,
			},
		},
		App: &ortb_V2_4.App{
			Id: &appID,
		},
		Imp: []*ortb_V2_4.Imp{
			{
				BidFloor: &bidFloor,
				Banner: &ortb_V2_4.Banner{
					W: &bannerWidth,
					H: &bannerHeight,
				},
			},
		},
	}
}

func (suite *FilterTestSuite) createTestBidRequestV25(country string, bidFloor float32, bannerWidth, bannerHeight int32, deviceIP string) *ortb_V2_5.BidRequest {
	return &ortb_V2_5.BidRequest{
		Device: &ortb_V2_5.Device{
			Ip: &deviceIP,
			Geo: &ortb_V2_5.Geo{
				Country: &country,
			},
		},
		Imp: []*ortb_V2_5.Imp{
			{
				BidFloor: &bidFloor,
				Banner: &ortb_V2_5.Banner{
					W: &bannerWidth,
					H: &bannerHeight,
				},
			},
		},
	}
}

func (suite *FilterTestSuite) createTestBidResponseV24(price float32, bidID, adID, impID string, hasBids bool) *ortb_V2_4.BidResponse {
	var seatbid *ortb_V2_4.SeatBid
	if hasBids {
		seatbid = &ortb_V2_4.SeatBid{
			Bid: []*ortb_V2_4.Bid{
				{
					Price: &price,
					Id:    &bidID,
					Adid:  &adID,
					Impid: &impID,
				},
			},
		}
	}

	return &ortb_V2_4.BidResponse{
		Seatbid: seatbid,
	}
}

func (suite *FilterTestSuite) createTestBidResponseV25(price float32, bidID, adID, impID string, hasBids bool) *ortb_V2_5.BidResponse {
	var seatbid *ortb_V2_5.SeatBid
	if hasBids {
		seatbid = &ortb_V2_5.SeatBid{
			Bid: []*ortb_V2_5.Bid{
				{
					Price: &price,
					Id:    &bidID,
					Adid:  &adID,
					Impid: &impID,
				},
			},
		}
	}

	return &ortb_V2_5.BidResponse{
		Seatbid: seatbid,
	}
}
