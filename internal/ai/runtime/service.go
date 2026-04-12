package runtime

import (
	"context"

	runtimeapp "cs-agent/internal/ai/runtime/app"
)

var Service = newService()

func newService() *service {
	return &service{
		app: runtimeapp.NewService(),
	}
}

type service struct {
	app *runtimeapp.Service
}

func (s *service) Run(ctx context.Context, req Request) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Run(ctx, req)
}

func (s *service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Resume(ctx, req)
}
