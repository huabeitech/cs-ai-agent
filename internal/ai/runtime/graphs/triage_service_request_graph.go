package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/services"
)

type TriageServiceRequestInput struct {
	Goal              string `json:"goal"`
	ObservedIssue     string `json:"observedIssue"`
	NeedTicket        bool   `json:"needTicket"`
	NeedHumanHandoff  bool   `json:"needHumanHandoff"`
	AdditionalContext string `json:"additionalContext"`
}

type TriageServiceRequestResult struct {
	Analysis          AnalyzeConversationResult `json:"analysis"`
	TicketDraft       *PrepareTicketDraftResult `json:"ticketDraft,omitempty"`
	RecommendedAction string                    `json:"recommendedAction"`
	Ready             bool                      `json:"ready"`
}

type TriageServiceRequestGraph struct {
	conversation models.Conversation
}

func NewTriageServiceRequestGraph(conversation models.Conversation) *TriageServiceRequestGraph {
	return &TriageServiceRequestGraph{conversation: conversation}
}

func (g *TriageServiceRequestGraph) Run(_ context.Context, argumentsInJSON string) (string, error) {
	input, err := g.parseInput(argumentsInJSON)
	if err != nil {
		return "", err
	}
	messages, _, _ := services.MessageService.FindByConversationIDCursor(g.conversation.ID, 0, 8, "", "")
	analysis := buildAnalyzeConversationResult(g.conversation, messages, AnalyzeConversationInput{
		Goal:              input.Goal,
		ObservedIssue:     input.ObservedIssue,
		NeedTicket:        input.NeedTicket,
		NeedHumanHandoff:  input.NeedHumanHandoff,
		AdditionalContext: input.AdditionalContext,
	})
	result := TriageServiceRequestResult{
		Analysis:          analysis,
		RecommendedAction: analysis.RecommendedNextAction,
		Ready:             analysis.RecommendedNextAction == "continue_answering" || analysis.RecommendedNextAction == "handoff_to_human",
	}
	if analysis.RecommendedNextAction == "prepare_ticket" {
		draft := buildPrepareTicketDraftResult(g.conversation, messages, PrepareTicketDraftInput{
			Issue: input.ObservedIssue,
		})
		result.TicketDraft = &draft
		result.Ready = draft.Ready
	}
	buf, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (g *TriageServiceRequestGraph) parseInput(argumentsInJSON string) (TriageServiceRequestInput, error) {
	var input TriageServiceRequestInput
	if strings.TrimSpace(argumentsInJSON) == "" {
		return input, nil
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return input, fmt.Errorf("invalid triage service request arguments: %w", err)
	}
	input.Goal = strings.TrimSpace(input.Goal)
	input.ObservedIssue = strings.TrimSpace(input.ObservedIssue)
	input.AdditionalContext = strings.TrimSpace(input.AdditionalContext)
	return input, nil
}
