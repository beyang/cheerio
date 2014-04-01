package cheerio

import (
	"fmt"
	"os"
	"strings"
)

type Requirement struct {
	Name       string
	Constraint string
	Version    string
}

// Parse requirements from a raw string in the requirements format expected by pip (e.g., in requirements.txt)
func ParseRequirements(rawReqs string) ([]*Requirement, error) {
	rawReqs = strings.TrimSpace(rawReqs)

	reqStrs := strings.Split(rawReqs, "\n")
	reqs := make([]*Requirement, 0)
	for _, reqStr := range reqStrs {
		if reqStr == "" {
			continue
		}

		if req, err := ParseRequirement(reqStr); err == nil {
			reqs = append(reqs, req)
		} else if reqHeaderRegexp.MatchString(reqStr) {
			// do nothing
		} else {
			os.Stderr.WriteString(fmt.Sprintf("[req] Could not parse requirement: %s\n", err))
		}
	}
	return reqs, nil
}

// Parse a single raw requirement, e.g., from "flask=1.0.1"
func ParseRequirement(reqStr string) (*Requirement, error) {
	reqStr = strings.TrimSpace(reqStr)
	match := requirementRegexp.FindStringSubmatch(reqStr)
	if len(match) != 5 {
		return nil, fmt.Errorf("Expected match of length 5, but got %+v from '%s'", match, reqStr)
	} else if match[0] != reqStr {
		return nil, fmt.Errorf("Unable to parse requirement from string: '%s'", reqStr)
	}
	return &Requirement{
		Name:       match[1],
		Constraint: match[3],
		Version:    match[4],
	}, nil
}

// Normalizes package names so they are comparable
func NormalizedPkgName(pkg string) string {
	return strings.ToLower(pkg)
}
