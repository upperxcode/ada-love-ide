package reasoning

import "strings"

type Type string

const (
	Plan    Type = "plan"
	Explore Type = "explore"
	Exec    Type = "exec"
	Read    Type = "read"
	Diff    Type = "diff"
	Text    Type = "text"
)

type Parser struct {
	currentType Type
	buffer      strings.Builder
}

func NewParser() *Parser {
	return &Parser{currentType: Plan}
}

func (rp *Parser) Feed(delta string) (Type, bool) {
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

func (rp *Parser) detect(haystack string) Type {
	if strings.Contains(haystack, "run") ||
		strings.Contains(haystack, "execute") ||
		strings.Contains(haystack, "command") ||
		strings.Contains(haystack, "go test") ||
		strings.Contains(haystack, "shell script") ||
		strings.Contains(haystack, "terminal") ||
		strings.Contains(haystack, "tool use") {
		return Exec
	}

	if strings.Contains(haystack, "search") ||
		strings.Contains(haystack, "find") ||
		strings.Contains(haystack, "explore") ||
		strings.Contains(haystack, "look for") ||
		strings.Contains(haystack, "list all") ||
		strings.Contains(haystack, "glob") ||
		strings.Contains(haystack, "grep") {
		return Explore
	}

	if strings.Contains(haystack, "read") ||
		strings.Contains(haystack, "open") ||
		strings.Contains(haystack, "view") ||
		strings.Contains(haystack, "check file") ||
		strings.Contains(haystack, "analyze") {
		return Read
	}

	if strings.Contains(haystack, "edit") ||
		strings.Contains(haystack, "write") ||
		strings.Contains(haystack, "change") ||
		strings.Contains(haystack, "modif") ||
		strings.Contains(haystack, "patch") ||
		strings.Contains(haystack, "create") {
		return Diff
	}

	if strings.Contains(haystack, "step") ||
		strings.Contains(haystack, "first") ||
		strings.Contains(haystack, "then") ||
		strings.Contains(haystack, "next") ||
		strings.Contains(haystack, "approach") ||
		strings.Contains(haystack, "strategy") ||
		strings.Contains(haystack, "plan") {
		return Plan
	}

	return rp.currentType
}
