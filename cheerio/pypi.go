package cheerio

import (
	"fmt"
	"github.com/beyang/cheerio/util"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

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

func (p *PackageIndex) PackageRequirements(pkg string) ([]*Requirement, error) {
	files, err := p.pkgFiles(pkg)
	if err != nil {
		return nil, err
	} else if len(files) == 0 {
		// os.Stderr.WriteString(fmt.Sprintf("[no-files] no files found for pkg %s\n", pkg))
		return nil, nil
	}

	if path := lastTar(files); path != "" {
		return p.fetchRequiresTar(path)
	} else if path := lastEgg(files); path != "" {
		return p.fetchRequiresZip(path, true)
	} else if path := lastZip(files); path != "" {
		return p.fetchRequiresZip(path, false)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("[tar/zip] no tar or zip found in %+v for pkg %s\n", files, pkg))
		return nil, nil
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

func (p *PackageIndex) fetchRequiresZip(path string, isEgg bool) ([]*Requirement, error) {
	uri := fmt.Sprintf("%s%s", p.URI, path)
	var pattern string
	if isEgg {
		pattern = "EGG-INFO/requires.txt"
	} else {
		pattern = "**/*.egg-info/requires.txt"
	}

	b, err := util.RemoteDecompress(uri, pattern, util.Zip)
	if err != nil {
		return nil, err
	}
	return ParseRequirements(string(b))
}

func (p *PackageIndex) fetchRequiresTar(path string) ([]*Requirement, error) {
	uri := fmt.Sprintf("%s%s", p.URI, path)
	b, err := util.RemoteDecompress(uri, "**/*.egg-info/requires.txt", util.Tar)
	if err != nil {
		return nil, err
	}
	return ParseRequirements(string(b))
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
