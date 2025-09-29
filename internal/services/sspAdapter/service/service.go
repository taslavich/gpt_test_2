package sppAdapter

import (
	"log"

	orchestrator "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/services/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ ISspAdapter = (*TSppAdapter)(nil)

func NewSspAdapter(
	addressOfOrchestrator string,
) *TSppAdapter {
	return &TSppAdapter{
		addressOfOrchestrator: addressOfOrchestrator,
	}
}

func (s *TSppAdapter) GetGrpClient() (
	orchestrator.OrchestratorServiceClient,
	func() error,
) {
	conn, err := grpc.NewClient(
		s.addressOfOrchestrator,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := orchestrator.NewOrchestratorServiceClient(conn)

	return client, conn.Close
}
