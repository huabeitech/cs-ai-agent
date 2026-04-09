package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
)

type CustomerServiceAgent struct {
	Inner adk.Agent
}

var _ adk.ResumableAgent = (*CustomerServiceAgent)(nil)

func (a *CustomerServiceAgent) Name(ctx context.Context) string {
	if a == nil || a.Inner == nil {
		return "customer_service_agent"
	}
	return a.Inner.Name(ctx)
}

func (a *CustomerServiceAgent) Description(ctx context.Context) string {
	if a == nil || a.Inner == nil {
		return "customer service chat agent"
	}
	return a.Inner.Description(ctx)
}

func (a *CustomerServiceAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	if a == nil || a.Inner == nil {
		iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
		gen.Send(&adk.AgentEvent{Err: context.Canceled})
		gen.Close()
		return iter
	}
	return a.Inner.Run(ctx, input, options...)
}

func (a *CustomerServiceAgent) Resume(ctx context.Context, info *adk.ResumeInfo, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	if a == nil || a.Inner == nil {
		iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
		gen.Send(&adk.AgentEvent{Err: fmt.Errorf("customer service agent is not initialized")})
		gen.Close()
		return iter
	}
	ra, ok := a.Inner.(adk.ResumableAgent)
	if !ok {
		iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
		gen.Send(&adk.AgentEvent{Err: fmt.Errorf("inner agent %q does not implement resumable agent", a.Inner.Name(ctx))})
		gen.Close()
		return iter
	}
	return ra.Resume(ctx, info, options...)
}
