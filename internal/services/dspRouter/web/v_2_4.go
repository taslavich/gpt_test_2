package dspRouterWeb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/constants"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/filter"
	dspRouterGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/dspRouter"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	utils "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/utils_grpc"
)

type DspMetaData struct {
	DspEndpoint string
	Code        int
	ErrMsg      string
}

type Server struct {
	ruleManager *filter.RuleManager

	fileLoader *filter.FileRuleLoader

	dspConfigPath string
	sppConfigPath string

	processor *filter.FilterProcessor

	reloadMutex  sync.RWMutex
	requestMutex sync.RWMutex

	dspEndpoints []string

	redisClient *redis.Client

	timeout time.Duration

	dspRouterGrpc.UnimplementedDspRouterServiceServer
}

func NewServer(
	ruleManager *filter.RuleManager,
	fileLoader *filter.FileRuleLoader,
	processor *filter.FilterProcessor,
	dspConfigPath string,
	sppConfigPath string,
	dspEndpoints []string,
	redisClient *redis.Client,
	timeout time.Duration,

) *Server {
	return &Server{
		ruleManager:   ruleManager,
		fileLoader:    fileLoader,
		processor:     processor,
		dspConfigPath: dspConfigPath,
		sppConfigPath: sppConfigPath,
		dspEndpoints:  dspEndpoints,
		redisClient:   redisClient,
		timeout:       timeout,
	}
}

func (s *Server) GetBids_V2_4(
	ctx context.Context,
	req *dspRouterGrpc.DspRouterRequest_V2_4,
) (
	*dspRouterGrpc.DspRouterResponse_V2_4,
	error,
) {

	var bdmu sync.Mutex
	var wg sync.WaitGroup

	dspEndpointLen := len(s.dspEndpoints)
	responsesCh := make(chan *ortb_V2_4.BidResponse, dspEndpointLen)
	dspMetaDataCh := make(chan *DspMetaData, dspEndpointLen)

	for i := range s.dspEndpoints {
		wg.Add(1)
		endpoint := s.dspEndpoints[i]
		go func(
			mu *sync.Mutex,
			req *dspRouterGrpc.DspRouterRequest_V2_4,
			dspEndpoint string,
		) {
			defer wg.Done()
			s.requestMutex.RLock()
			filterResult := s.processor.ProcessRequestForDSPV24(dspEndpoint, req.BidRequest)
			s.requestMutex.RUnlock()

			if !filterResult.Allowed {
				return
			}

			resp, code, errMsg := s.getBidsFromDSPbyHTTP_V_2_4(req, dspEndpoint, req.SppEndpoint)

			dspMetaDataCh <- &DspMetaData{
				DspEndpoint: dspEndpoint,
				Code:        code,
				ErrMsg:      errMsg,
			}

			if filterRes := s.processor.ProcessResponseForSPPV24(req.SppEndpoint, resp); !filterRes.Allowed {
				return
			}

			if resp != nil {
				responsesCh <- resp
			}
		}(
			&bdmu,
			req,
			endpoint,
		)
	}

	wg.Wait()
	close(responsesCh)
	close(dspMetaDataCh)

	dspMetaData := make([]*DspMetaData, 0)
	for d := range dspMetaDataCh {
		dspMetaData = append(dspMetaData, d)
	}

	bidRespsData, err := json.Marshal(dspMetaData)
	if err != nil {
		fmt.Printf("failed to marshal slice in GetBids_V2_4: %w", err)
	}

	if err := utils.WriteJsonToRedis(ctx, s.redisClient, req.GlobalId, constants.BID_RESPONSES_COLUMN, bidRespsData); err != nil {
		fmt.Printf("failed to WriteJsonToRedis Bid Responses in GetBids_V2_4: %v", err)
	}

	return &dspRouterGrpc.DspRouterResponse_V2_4{
		BidRequest: req.BidRequest,
		BidResponses: func() []*ortb_V2_4.BidResponse {
			responses := make([]*ortb_V2_4.BidResponse, 0)
			for resp := range responsesCh {
				responses = append(responses, resp)
			}

			return responses
		}(),
		GlobalId: req.GlobalId,
	}, nil
}

func (s *Server) getBidsFromDSPbyHTTP_V_2_4(req *dspRouterGrpc.DspRouterRequest_V2_4, dspEndpoint, sppEndpoint string) (
	br *ortb_V2_4.BidResponse,
	code int,
	errMsg string,
) {
	jsonData, err := json.Marshal(req.BidRequest)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not marshal in GetBids_V2_4: %w", err)
	}
	client := &http.Client{Timeout: s.timeout}

	resp, err := client.Post(
		dspEndpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not post req to dsps in GetBids_V2_4: %w", err)
	}
	defer func() {
		if retErr := resp.Body.Close(); err != nil {
			errMsg = fmt.Sprintf(
				"Cannot close resp in GetBids_V2_4: %w",
				retErr,
			)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not read body to dsps in GetBids_V2_4: %w", err)
	}

	var grpcResp *ortb_V2_4.BidResponse
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		grpcResp = &ortb_V2_4.BidResponse{}
		if err := json.Unmarshal(body, grpcResp); err != nil {
			return nil,
				resp.StatusCode,
				fmt.Sprintf("Can not unmarshal body from dsps in GetBids_V2_4: %w", err)
		}
	}

	return grpcResp,
		resp.StatusCode,
		"NULL"
}

/*func (h *Server) GetDSPRules_V2_4(ctx context.Context, req *dspRouterGrpc.GetRulesRequest) (*dspRouterGrpc.JsonResponse, error) {
	h.reloadMutex.RLock()
	defer h.reloadMutex.RUnlock()

	jsonData, err := os.ReadFile(h.dspConfigPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error reading config file: %v", err)
	}

	var config filter.SimpleRuleConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, status.Errorf(codes.Internal, "Error parsing config: %v", err)
	}

	dspOnlyConfig := map[string]interface{}{
		"version": config.Version,
		"dsps":    config.DSPs,
	}

	dspOnlyJson, err := json.Marshal(dspOnlyConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error marshaling DSP config: %v", err)
	}

	return &dspRouterGrpc.JsonResponse{
		JsonData: dspOnlyJson,
	}, nil
}

func (h *Server) GetSPPRules_V2_4(ctx context.Context, req *dspRouterGrpc.GetRulesRequest) (*dspRouterGrpc.JsonResponse, error) {
	h.reloadMutex.RLock()
	defer h.reloadMutex.RUnlock()

	jsonData, err := os.ReadFile(h.sppConfigPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error reading config file: %v", err)
	}

	var config filter.SimpleRuleConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, status.Errorf(codes.Internal, "Error parsing config: %v", err)
	}

	sppOnlyConfig := map[string]interface{}{
		"version": config.Version,
		"spps":    config.SPPs,
	}

	sppOnlyJson, err := json.Marshal(sppOnlyConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error marshaling SPP config: %v", err)
	}

	return &dspRouterGrpc.JsonResponse{
		JsonData: sppOnlyJson,
	}, nil
}

func (h *Server) UpdateDSPRules_V2_4(ctx context.Context, req *dspRouterGrpc.JsonRequest) (*dspRouterGrpc.UpdateRulesResponse, error) {
	return h.updateRulesInternal(
		req.JsonData,
		h.dspConfigPath,
		filter.ValidateDSPConfig,
		h.fileLoader.LoadDSPRules,
	)
}

func (h *Server) UpdateSPPRules_V2_4(ctx context.Context, req *dspRouterGrpc.JsonRequest) (*dspRouterGrpc.UpdateRulesResponse, error) {
	return h.updateRulesInternal(
		req.JsonData,
		h.sppConfigPath,
		filter.ValidateSPPConfig,
		h.fileLoader.LoadSPPRules,
	)
}

func (h *Server) updateRulesInternal(
	jsonData []byte,
	configPath string,
	validateFunc func(*filter.SimpleRuleConfig) error,
	loadFunc func() error,
) (*dspRouterGrpc.UpdateRulesResponse, error) {

	h.requestMutex.Lock()
	defer h.requestMutex.Unlock()

	h.reloadMutex.Lock()
	defer h.reloadMutex.Unlock()

	var config filter.SimpleRuleConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return &dspRouterGrpc.UpdateRulesResponse{
			Success: false,
			Message: "Invalid JSON format: " + err.Error(),
		}, nil
	}

	if err := validateFunc(&config); err != nil {
		return &dspRouterGrpc.UpdateRulesResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		}, nil
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return &dspRouterGrpc.UpdateRulesResponse{
			Success: false,
			Message: "Error writing config: " + err.Error(),
		}, nil
	}

	if err := loadFunc(); err != nil {
		return &dspRouterGrpc.UpdateRulesResponse{
			Success: false,
			Message: "Error loading new rules: " + err.Error(),
		}, nil
	}

	return &dspRouterGrpc.UpdateRulesResponse{
		Success: true,
		Message: "Rules updated successfully",
	}, nil
}
*/
