package cheerio

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beyang/cheerio/util"
	"github.com/beyang/go-version"
)

var DefaultPyPI = &PackageIndex{URI: "https://pypi.python.org"}

type PackageIndex struct {
	URI string
}

// Get names of all packages served by a PyPI server.
func (p *PackageIndex) AllPackages() ([]string, error) {
	pkgs := make([]string, 0)

	resp, err := http.Get(fmt.Sprintf("%s/simple", p.URI))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	matches := allPkgRegexp.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		if len(match) != 3 {
			return nil, fmt.Errorf("Unexpected number of submatches: %d, %v", len(match), match)
		} else if match[1] != match[2] {
			return nil, fmt.Errorf("Names do not match %s != %s", match[1], match[2])
		} else {
			pkgs = append(pkgs, match[1])
		}
	}

	return pkgs, nil
}

var requiresTxtTarPattern = regexp.MustCompile(`(?:[^/]+/)*(?:[^/]*\.egg\-info/requires\.txt)`)
var requiresTxtEggPattern = regexp.MustCompile(`EGG\-INFO/requires\.txt`)
var requiresTxtZipPattern = requiresTxtTarPattern

// Fetches package requirements from PyPI by downloading the package archive and extracting the requires.txt file.  If no such file exists (sometimes
// it doesn't), returns an error.
func (p *PackageIndex) FetchPackageRequirements(pkg string) ([]*Requirement, error) {
	b, err := p.FetchRawMetadata(pkg, requiresTxtTarPattern, requiresTxtEggPattern, requiresTxtZipPattern)
	if err != nil {
		if strings.Contains(err.Error(), "[no-files]") { // may not have a requires.txt
			return nil, nil
		} else {
			return nil, err
		}
	}
	return ParseRequirements(string(b))
}

func (p *PackageIndex) FetchRawMetadata(pkg string, tarPattern, eggPattern, zipPattern *regexp.Regexp) ([]byte, error) {
	files, err := p.pkgFiles(pkg)
	if err != nil {
		return nil, err
	} else if len(files) == 0 {
		return nil, fmt.Errorf("[no-files] no files found for pkg %s", pkg)
	}

	// Sort files in version order
	version.Sort(files)

	// Get the latest version
	if path := lastTar(files); path != "" {
		return util.RemoteDecompress(fmt.Sprintf("%s%s", p.URI, path), tarPattern, util.Tar)
	} else if path := lastEgg(files); path != "" {
		return util.RemoteDecompress(fmt.Sprintf("%s%s", p.URI, path), eggPattern, util.Zip)
	} else if path := lastZip(files); path != "" {
		return util.RemoteDecompress(fmt.Sprintf("%s%s", p.URI, path), zipPattern, util.Zip)
	} else {
		return nil, fmt.Errorf("[tar/zip] no tar or zip found in %+v for pkg %s", files, pkg)
	}
}

var allPkgRegexp = regexp.MustCompile(`<a href='([A-Za-z0-9\._\-]+)'>([A-Za-z0-9\._\-]+)</a><br/>`)
var pkgFilesRegexp = regexp.MustCompile(`<a href="([/A-Za-z0-9\._\-]+)#md5=[0-9a-z]+"[^>]*>([A-Za-z0-9\._\-]+)</a><br/>`)
var tarRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.(?:tar\.(?:gz|bz2)|tgz)`)
var zipRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.zip`)
var eggRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.egg`)
var requirementRegexp = regexp.MustCompile(`(?P<package>[A-Za-z0-9\._\-]+)(?:\[([A-Za-z0-9\._\-]+)\])?\s*(?:(?P<constraint>==|>=|>|<|<=)\s*(?P<version>[A-Za-z0-9\._\-]+)(?:\s*,\s*[<>=!]+\s*[a-z0-9\.]+)?)?`)
var reqHeaderRegexp = regexp.MustCompile(`\[[A-Za-z0-9\._\-]+\]`)

// Helpers

func (p *PackageIndex) pkgFiles(pkg string) ([]string, error) {
	files := make([]string, 0)

	uriPath := fmt.Sprintf("/simple/%s", pkg)
	uri := fmt.Sprintf("%s%s", p.URI, uriPath)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	matches := pkgFilesRegexp.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		if len(match) != 3 {
			return nil, fmt.Errorf("Unexpected number of submatches: %d, %v", len(match), match)
		} else {
			files = append(files, filepath.Clean(filepath.Join(uriPath, match[1])))
		}
	}

	return files, nil
}
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
