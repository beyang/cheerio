package cheerio

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

// Return requirements for python PyPI package in directory
func RequirementsForDir(dir string) ([]*Requirement, error) {
	reqs := make(map[string]*Requirement)

	// If this contains a PyPI module, get requirements from PyPI graph
	if pyPIName := pypiNameFromRepoDir(dir); pyPIName != "" {
		requires := DefaultPyPIGraph.Requires(pyPIName)
		for _, req := range requires {
			reqs[NormalizedPkgName(req)] = &Requirement{Name: req}
		}
	}

	// If repo contains requirements.txt, parse requirements from that (these should be more specific than those contained in a PyPIGraph, because
	// they will often include version info).
	reqFile := filepath.Join(dir, "requirements.txt")
	if reqFileContents, err := ioutil.ReadFile(reqFile); err == nil {
		if rawReqs, err := ParseRequirements(string(reqFileContents)); err == nil {
			// Note: this currently doesn't handle pip+git-URL requirements
			for _, rawReq := range rawReqs {
				reqs[NormalizedPkgName(rawReq.Name)] = rawReq
			}
		}
	}

	// if len(requirements) == 0 {
	// 	// TODO: use depdump.py to best-effort get requirements
	// }

	reqList := make([]*Requirement, 0)
	for _, req := range reqs {
		reqList = append(reqList, req)
	}
	return reqList, nil
}

var setupNameRegexp = regexp.MustCompile(`name\s?=\s?['"](?P<name>[A-Za-z0-9\._\-]+)['"]`)

func pypiNameFromRepoDir(dir string) string {
	setupFile := filepath.Join(dir, "setup.py")
	setupBytes, err := ioutil.ReadFile(setupFile)
	if err != nil {
		return ""
	}
	matches := setupNameRegexp.FindAllStringSubmatch(string(setupBytes), -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[0][1]
}
