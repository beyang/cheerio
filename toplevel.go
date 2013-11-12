package cheerio

import (
	"regexp"
	"strings"
)

var topLevelTxtPattern = regexp.MustCompile(`(?:[^/]+/)*(?:[^/]*\.egg\-info/top_level\.txt)`)

func (p *PackageIndex) FetchSourceTopLevelModules(pkg string) ([]string, error) {
	b, err := p.FetchRawMetadata(pkg, topLevelTxtPattern, topLevelTxtPattern, topLevelTxtPattern)
	if err != nil {
		// If error, try to fall back to hard-coded top-level modules
		if hardCodedModules, in := pypiTopLevelModules[pkg]; in {
			return hardCodedModules, nil
		} else {
			return nil, err
		}
	}

	var modules []string
	for _, line := range strings.Split(string(b), "\n") {
		if module := strings.TrimSpace(line); module != "" {
			modules = append(modules, module)
		}
	}
	return modules, nil
}

var pypiTopLevelModules = map[string][]string{
	"pyyaml":          []string{"yaml"},
	"django-tastypie": []string{"tastypie"},
	"twisted":         []string{"twisted"},
}
