package factory

import (
	"context"

	einostore "cs-agent/internal/ai/runtime/internal/impl/store"

	"github.com/cloudwego/eino/adk"
)

type RunnerFactory struct{}

func NewRunnerFactory() *RunnerFactory {
	return &RunnerFactory{}
}

func (f *RunnerFactory) Build(ctx context.Context, agent adk.Agent, enableStreaming bool, enableCheckpoint bool) *adk.Runner {
	var checkpointStore adk.CheckPointStore
	if enableCheckpoint {
		checkpointStore = einostore.DefaultCheckPointStore
	}
	return adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: enableStreaming,
		CheckPointStore: checkpointStore,
	})
}
