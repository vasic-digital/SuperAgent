package main

import (
	"context"
	llm "github.com/superagent/superagent/internal/llm"
	models "github.com/superagent/superagent/internal/models"
	pb "github.com/superagent/superagent/pkg/api"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type grpcServer struct{}

func (s *grpcServer) Complete(ctx context.Context, req *pb.CompletionRequest) (*pb.CompletionResponse, error) {
	internal := &models.LLMRequest{
		ID:             "",
		SessionID:      req.SessionID,
		UserID:         req.UserID,
		Prompt:         req.Prompt,
		MemoryEnhanced: req.MemoryEnhanced,
		Memory:         nil,
		ModelParams:    nil,
		EnsembleConfig: nil,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	responses, selected, err := llm.RunEnsemble(internal)
	if err != nil {
		return &pb.CompletionResponse{Response: "", Confidence: 0}, err
	}
	var out pb.CompletionResponse
	if len(responses) > 0 && responses[0] != nil {
		out.Response = responses[0].Content
		out.Confidence = responses[0].Confidence
		out.ProviderName = responses[0].ProviderName
	}
	if selected != nil {
		out.Response = selected.Content
		out.Confidence = selected.Confidence
	}
	return &out, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// pb.RegisterLLMFacadeServer(s, &grpcServer{}) // enable when pb.go is generated
	log.Println("gRPC-like server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
