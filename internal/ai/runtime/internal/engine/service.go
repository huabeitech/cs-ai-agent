package engine

import (
	"context"

	"cs-agent/internal/ai/runtime/internal/executor"
)

type Service struct {
	executor *executor.Service
}

func NewService() *Service {
	return &Service{
		executor: executor.NewService(),
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	return s.ExecuteRun(ctx, req)
}

func (s *Service) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	return s.executor.ExecuteRun(ctx, executor.RunInput(req))
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	return s.ExecuteResume(ctx, req)
}

func (s *Service) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	return s.executor.ExecuteResume(ctx, executor.ResumeInput(req))
}
