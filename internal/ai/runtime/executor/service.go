package executor

import (
	"context"

	internalexecutor "cs-agent/internal/ai/runtime/internal/executor"
)

type Service struct {
	inner *internalexecutor.Service
}

func NewService() *Service {
	return &Service{
		inner: internalexecutor.NewService(),
	}
}

func (s *Service) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteRun(ctx, req)
}

func (s *Service) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteResume(ctx, req)
}
