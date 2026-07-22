package adapters

import (
	"strings"
)

type ReasoningType string

const (
	ReasoningPlan    ReasoningType = "plan"
	ReasoningExplore ReasoningType = "explore"
	ReasoningExec    ReasoningType = "exec"
	ReasoningRead    ReasoningType = "read"
	ReasoningDiff    ReasoningType = "diff"
	ReasoningText    ReasoningType = "text"
)

type ReasoningParser struct {
	currentType ReasoningType
	buffer      strings.Builder
}

func NewReasoningParser() *ReasoningParser {
	return &ReasoningParser{
		currentType: ReasoningPlan,
	}
}

func (rp *ReasoningParser) Feed(delta string) (ReasoningType, bool) {
	rp.buffer.WriteString(delta)

	if rp.buffer.Len() > 500 {
		s := rp.buffer.String()
		keep := 200
		if len(s) > keep {
			s = s[len(s)-keep:]
		}
		rp.buffer.Reset()
		rp.buffer.WriteString(s)
	}

	haystack := strings.ToLower(rp.buffer.String())

	detected := rp.detect(haystack)
	changed := detected != rp.currentType
	if changed {
		rp.currentType = detected
	}
	return detected, changed
}

func (rp *ReasoningParser) detect(haystack string) ReasoningType {
	if strings.Contains(haystack, "run") ||
		strings.Contains(haystack, "execute") ||
		strings.Contains(haystack, "command") ||
		strings.Contains(haystack, "go test") ||
		strings.Contains(haystack, "shell script") ||
		strings.Contains(haystack, "terminal") ||
		strings.Contains(haystack, "tool use") {
		return ReasoningExec
	}

	if strings.Contains(haystack, "search") ||
		strings.Contains(haystack, "find") ||
		strings.Contains(haystack, "explore") ||
		strings.Contains(haystack, "look for") ||
		strings.Contains(haystack, "list all") ||
		strings.Contains(haystack, "glob") ||
		strings.Contains(haystack, "grep") {
		return ReasoningExplore
	}

	if strings.Contains(haystack, "read") ||
		strings.Contains(haystack, "open") ||
		strings.Contains(haystack, "view") ||
		strings.Contains(haystack, "check file") ||
		strings.Contains(haystack, "analyze") {
		return ReasoningRead
	}

	if strings.Contains(haystack, "edit") ||
		strings.Contains(haystack, "write") ||
		strings.Contains(haystack, "change") ||
		strings.Contains(haystack, "modif") ||
		strings.Contains(haystack, "patch") ||
		strings.Contains(haystack, "create") {
		return ReasoningDiff
	}

	if strings.Contains(haystack, "step") ||
		strings.Contains(haystack, "first") ||
		strings.Contains(haystack, "then") ||
		strings.Contains(haystack, "next") ||
		strings.Contains(haystack, "approach") ||
		strings.Contains(haystack, "strategy") ||
		strings.Contains(haystack, "plan") {
		return ReasoningPlan
	}

	return rp.currentType
}
