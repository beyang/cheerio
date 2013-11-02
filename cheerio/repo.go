package cheerio

import (
	"fmt"
	"regexp"
)

var repoPatterns = []*regexp.Regexp{
	regexp.MustCompile(`Home-page: (https?://github.com/(:?[^/\n]+)/(:?[^/\n]+))(:?/.*)?\n`),
	regexp.MustCompile(`Home-page: (https?://bitbucket.org/(:?[^/\n]+)/(:?[^/\n]+))(:?/.*)?\n`),
	regexp.MustCompile(`Home-page: (https?://code.google.com/p/(:?[^/\n]+))(:?/.*)?\n`),
}

var homepageRegexp = regexp.MustCompile(`Home-page: (.+)\n`)

func (p *PackageIndex) FetchSourceRepoURI(pkg string) (string, error) {
	pattern := "**/PKG-INFO"
	b, err := p.FetchRawMetadata(pkg, pattern, pattern, pattern)
	if err != nil {
		return "", nil
	}
	rawMetadata := string(b)

	for _, pattern := range repoPatterns {
		if match := pattern.FindStringSubmatch(rawMetadata); len(match) >= 1 {
			return match[1], nil
		}
	}

	if match := homepageRegexp.FindStringSubmatch(rawMetadata); len(match) >= 1 {
		return "", fmt.Errorf("Could not parse repo URI from homepage: %s", match[1])
	}
	return "", fmt.Errorf("No homepage found in metadata: %s", rawMetadata)
}
