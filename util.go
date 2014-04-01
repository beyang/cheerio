package cheerio

import (
	"regexp"
	"strings"
)

// Normalizes package names so they are comparable
func NormalizedPkgName(pkg string) string {
	return strings.ToLower(pkg)
}

// Convenience functions that get the last instance of a type of file
var tarRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.(?:tar\.(?:gz|bz2)|tgz)`)
var zipRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.zip`)
var eggRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.egg`)

func lastTar(files []string) string {
	for f := len(files) - 1; f >= 0; f-- {
		if tarRegexp.MatchString(files[f]) {
			return files[f]
		}
	}
	return ""
}

func lastEgg(files []string) string {
	for f := len(files) - 1; f >= 0; f-- {
		if eggRegexp.MatchString(files[f]) {
			return files[f]
		}
	}
	return ""
}

func lastZip(files []string) string {
	for f := len(files) - 1; f >= 0; f-- {
		if zipRegexp.MatchString(files[f]) {
			return files[f]
		}
	}
	return ""
}
