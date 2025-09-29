package orchestrator

import (
	bidEngineGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/bidEngine"
	dspRouterGrpc "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/dspRouter"
)

var _ IOrchestrator = (*TOrchestrator)(nil)

type GrpcClients struct {
	BidEngineGrpcClient bidEngineGrpc.BidEngineServiceClient
	DspRouterGrpcClient dspRouterGrpc.DspRouterServiceClient
}

type TOrchestrator struct {
	addressOfBidEngine string
	addressOfDspRouter string
}

type IOrchestrator interface {
	GetGrpClients() (*GrpcClients, func())
}
