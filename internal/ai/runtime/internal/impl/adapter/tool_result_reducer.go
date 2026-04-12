package adapter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cs-agent/internal/ai/mcps"
)

const (
	maxToolResultSummaryChars = 4000
	maxToolResultSegments     = 12
)

var reductionInfoPattern = regexp.MustCompile(`\[tool result reduced: original_length=(\d+), kept_length=(\d+)\]`)

type ReductionInfo struct {
	Reduced       bool
	OriginalChars int
	KeptChars     int
}

// BuildReducedToolResultSummary returns a bounded text summary for MCP tool results.
// It keeps the main payload visible to the model while preventing a single large tool
// response from exhausting too much context.
func BuildReducedToolResultSummary(result *mcps.ToolCallResult) string {
	if result == nil {
		return ""
	}
	segments := collectToolResultSegments(result)
	if len(segments) == 0 {
		return ""
	}
	text := strings.TrimSpace(strings.Join(segments, "\n"))
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxToolResultSummaryChars {
		return text
	}
	truncated := strings.TrimSpace(string(runes[:maxToolResultSummaryChars]))
	return fmt.Sprintf("%s\n\n[tool result reduced: original_length=%d, kept_length=%d]", truncated, len(runes), maxToolResultSummaryChars)
}

func ParseReductionInfo(summary string) ReductionInfo {
	matches := reductionInfoPattern.FindStringSubmatch(strings.TrimSpace(summary))
	if len(matches) != 3 {
		return ReductionInfo{}
	}
	originalChars, err1 := strconv.Atoi(matches[1])
	keptChars, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return ReductionInfo{}
	}
	return ReductionInfo{
		Reduced:       true,
		OriginalChars: originalChars,
		KeptChars:     keptChars,
	}
}

func collectToolResultSegments(result *mcps.ToolCallResult) []string {
	segments := make([]string, 0, len(result.Content)+2)
	if result.IsError {
		segments = append(segments, "tool returned an error")
	}
	if result.StructuredContent != nil {
		if data, err := json.Marshal(result.StructuredContent); err == nil {
			segments = appendNonBlankSegment(segments, string(data))
		}
	}
	for _, item := range result.Content {
		if len(segments) >= maxToolResultSegments {
			segments = append(segments, "[tool result reduced: remaining segments omitted]")
			break
		}
		switch item.Type {
		case "text":
			segments = appendNonBlankSegment(segments, item.Text)
		default:
			if item.Data == nil {
				continue
			}
			if data, err := json.Marshal(item.Data); err == nil {
				segments = appendNonBlankSegment(segments, string(data))
			}
		}
	}
	return segments
}

func appendNonBlankSegment(input []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return input
	}
	return append(input, value)
}
