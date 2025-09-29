package dspRouterWeb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"gitlab.com/twinbid-exchange/RTB-exchange/internal/constants"
	dspRouterGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/dspRouter"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
	utils "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/utils_grpc"
)

func (s *Server) GetBids_V2_5(
	ctx context.Context,
	req *dspRouterGrpc.DspRouterRequest_V2_5,
) (
	*dspRouterGrpc.DspRouterResponse_V2_5,
	error,
) {
	originReq := req

	var bdmu sync.Mutex
	var wg sync.WaitGroup

	dspEndpointLen := len(s.dspEndpoints)
	responsesCh := make(chan *ortb_V2_5.BidResponse, dspEndpointLen)
	dspMetaDataCh := make(chan *DspMetaData, dspEndpointLen)

	for i := range s.dspEndpoints {
		wg.Add(1)
		endpoint := s.dspEndpoints[i]
		go func(
			mu *sync.Mutex,
			req *dspRouterGrpc.DspRouterRequest_V2_5,
			endpoint string,
		) {
			defer wg.Done()
			s.requestMutex.RLock()
			filterResult := s.processor.ProcessRequestForDSPV25(endpoint, req.BidRequest)
			s.requestMutex.RUnlock()

			if !filterResult.Allowed {
				return
			}

			resp, code, errMsg := s.getBidsFromDSPbyHTTP_V_2_5(req, endpoint)

			dspMetaDataCh <- &DspMetaData{
				DspEndpoint: endpoint,
				Code:        code,
				ErrMsg:      errMsg,
			}

			if filterRes := s.processor.ProcessResponseForSPPV25(req.SppEndpoint, resp); !filterRes.Allowed {
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
		fmt.Printf("failed to marshal slice in GetBids_V2_5: %w", err)
	}

	if err := utils.WriteJsonToRedis(ctx, s.redisClient, req.GlobalId, constants.BID_RESPONSES_COLUMN, bidRespsData); err != nil {
		fmt.Printf("failed to WriteJsonToRedis Bid Responses in GetBids_V2_5: %v", err)
	}

	return &dspRouterGrpc.DspRouterResponse_V2_5{
		BidRequest: originReq.BidRequest,
		BidResponses: func() []*ortb_V2_5.BidResponse {
			responses := make([]*ortb_V2_5.BidResponse, 0)
			for resp := range responsesCh {
				responses = append(responses, resp)
			}
			return responses
		}(),
		GlobalId: req.GlobalId,
	}, nil
}

func (s *Server) getBidsFromDSPbyHTTP_V_2_5(req *dspRouterGrpc.DspRouterRequest_V2_5, dspEndpoint string) (
	br *ortb_V2_5.BidResponse,
	code int,
	errMsg string,
) {
	client := &http.Client{ /*Timeout: s.timeout*/ }

	jsonData, err := json.Marshal(req.BidRequest)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not marshal in GetBids_V2_5: %w", err)
	}

	resp, err := client.Post(
		dspEndpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not post req to dsps in GetBids_V2_5: %w", err)
	}
	defer func() {
		if retErr := resp.Body.Close(); err != nil {
			errMsg = fmt.Sprintf(
				"Cannot close resp in GetBids_V2_5: %w",
				retErr,
			)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Sprintf("Can not read body to dsps in GetBids_V2_5: %w", err)
	}

	var grpcResp *ortb_V2_5.BidResponse
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		grpcResp = &ortb_V2_5.BidResponse{}
		if err := json.Unmarshal(body, grpcResp); err != nil {
			return nil,
				resp.StatusCode,
				fmt.Sprintf("Can not unmarshal body from dsps in GetBids_V2_5: %w", err)
		}
	}

	return grpcResp,
		resp.StatusCode,
		"NULL"
}
