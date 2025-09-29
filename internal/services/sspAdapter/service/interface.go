package sppAdapter

import (
	orchestrator "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/orchestrator"
)

type TSppAdapter struct {
	IsBadIp       func(ipStr string) (bool, error)
	GetCountryISO func(ipStr string) (string, error)

	addressOfOrchestrator string
}

type ISspAdapter interface {
	GetGrpClient() (
		orchestrator.OrchestratorServiceClient,
		func() error,
	)
}
