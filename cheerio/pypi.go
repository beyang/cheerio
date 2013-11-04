package cheerio

import (
	"fmt"
	"github.com/beyang/cheerio/util"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var DefaultPyPI = &PackageIndex{URI: "https://pypi.python.org"}

type PackageIndex struct {
	URI string
}

func (p *PackageIndex) AllPackages() ([]string, error) {
	pkgs := make([]string, 0)

	resp, err := http.Get(fmt.Sprintf("%s/simple", p.URI))
	if err != nil {
		return nil, err
	}
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

func (p *PackageIndex) FetchPackageRequirements(pkg string) ([]*Requirement, error) {
	tarPattern := "**/*.egg-info/requires.txt"
	eggPattern := "EGG-INFO/requires.txt"
	zipPattern := tarPattern

	b, err := p.FetchRawMetadata(pkg, tarPattern, eggPattern, zipPattern)
	if err != nil {
		if strings.Contains(err.Error(), "[no-files]") { // may not have a requires.txt
			return nil, nil
		} else {
			return nil, err
		}
	}
	return ParseRequirements(string(b))
}

func (p *PackageIndex) FetchRawMetadata(pkg string, tarPattern, eggPattern, zipPattern string) ([]byte, error) {
	files, err := p.pkgFiles(pkg)
	if err != nil {
		return nil, err
	} else if len(files) == 0 {
		return nil, fmt.Errorf("[no-files] no files found for pkg %s", pkg)
	}

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
