package cheerio

import (
	"strings"
)

func NormalizedPkgName(pkg string) string {
	return strings.ToLower(pkg)
}
