package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type PackageIndex struct {
	URI string
}

type Requirement struct {
	Name string
}

var allPkgRegexp = regexp.MustCompile(`<a href='([A-Za-z0-9\._\-]+)'>([A-Za-z0-9\._\-]+)</a><br/>`)
var pkgFilesRegexp = regexp.MustCompile(`<a href="([/A-Za-z0-9\._\-]+)#md5=[0-9a-z]+"[^>]*>([A-Za-z0-9\._\-]+)</a><br/>`)
var tarRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.(?:tar\.(?:gz|bz2)|tgz)`)
var zipRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.zip`)
var eggRegexp = regexp.MustCompile(`[/A-Za-z0-9\._\-]+\.egg`)
var requirementRegexp = regexp.MustCompile(`(?P<package>[A-Za-z0-9\._\-]+)(?:\[([A-Za-z0-9\._\-]+)\])?\s*(?:(?P<constraint>==|>=|>|<|<=)\s*(?P<version>[A-Za-z0-9\._\-]+)(?:\s*,\s*[<>=!]+\s*[a-z0-9\.]+)?)?`)
var reqHeaderRegexp = regexp.MustCompile(`\[[A-Za-z0-9\._\-]+\]`)

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
	f, err := ioutil.TempFile("", "pypigraph_zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if err := f.Close(); err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s%s", p.URI, path)
	wget := exec.Command("wget", uri, "-O", f.Name())
	var eggInfoFilePattern string
	if isEgg {
		eggInfoFilePattern = "EGG-INFO/requires.txt"
	} else {
		eggInfoFilePattern = "**/*.egg-info/requires.txt"
	}
	unzip := exec.Command("unzip", "-cq", f.Name(), eggInfoFilePattern)

	err = wget.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running wget: %s", err)
	}

	unzipOut, err := unzip.StdoutPipe()
	if err != nil {
		return nil, err
	}
	unzipErr, err := unzip.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = unzip.Start()
	if err != nil {
		return nil, err
	}
	var unzipErrput, unzipOutput []byte
	go func() {
		unzipErrput, _ = ioutil.ReadAll(unzipErr)
	}()
	go func() {
		unzipOutput, _ = ioutil.ReadAll(unzipOut)
	}()
	err = unzip.Wait()
	if err != nil {
		if strings.Contains(string(unzipErrput), "filename not matched:") {
			// os.Stderr.WriteString(fmt.Sprintf("[requires.txt] no requires.txt found in %s\n", uri))
			return nil, nil
		} else {
			return nil, fmt.Errorf("Error running unzip on file %s: %s, [%s]", f.Name(), err, string(unzipErrput))
		}
	}

	rawReqs := strings.TrimSpace(string(unzipOutput))
	return parseRequirements(rawReqs)
}

func (p *PackageIndex) fetchRequiresTar(path string) ([]*Requirement, error) {
	uri := fmt.Sprintf("%s%s", p.URI, path)
	curl := exec.Command("curl", uri)
	tar := exec.Command("tar", "-xvO", "--include", "**/*.egg-info/requires.txt")

	curlOut, err := curl.StdoutPipe()
	if err != nil {
		return nil, err
	}
	tarIn, err := tar.StdinPipe()
	if err != nil {
		return nil, err
	}
	tarOut, err := tar.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		io.Copy(tarIn, curlOut)
		tarIn.Close()
	}()

	curl.Start()
	tar.Start()

	tarOutput, err := ioutil.ReadAll(tarOut)
	if err != nil {
		return nil, err
	}

	curl.Wait()
	tar.Wait()

	rawReqs := strings.TrimSpace(string(tarOutput))
	if rawReqs == "" {
		// os.Stderr.WriteString(fmt.Sprintf("[requires.txt] no requires.txt found in %s\n", uri))
		return nil, nil
	}

	return parseRequirements(rawReqs)
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

func parseRequirements(rawReqs string) ([]*Requirement, error) {
	rawReqs = strings.TrimSpace(rawReqs)

	reqStrs := strings.Split(rawReqs, "\n")
	reqs := make([]*Requirement, 0)
	for _, reqStr := range reqStrs {
		if reqStr == "" {
			continue
		}

		if req, err := parseRequirement(reqStr); err == nil {
			reqs = append(reqs, req)
		} else if reqHeaderRegexp.MatchString(reqStr) {
			// do nothing
		} else {
			os.Stderr.WriteString(fmt.Sprintf("[req] Could not parse requirement: %s\n", err))
		}
	}
	return reqs, nil
}

func parseRequirement(reqStr string) (*Requirement, error) {
	reqStr = strings.TrimSpace(reqStr)
	match := requirementRegexp.FindStringSubmatch(reqStr)
	if len(match) != 5 {
		return nil, fmt.Errorf("Expected match of length 5, but got %+v from '%s'", match, reqStr)
	} else if match[0] != reqStr {
		return nil, fmt.Errorf("Unable to parse requirement from string: '%s'", reqStr)
	}

	return &Requirement{Name: match[1]}, nil
}

func main() {
	p := &PackageIndex{URI: "https://pypi.python.org"}
	pkgs, err := p.AllPackages()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[FATAL] %s\n", err))
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		reqs, err := p.PackageRequirements(pkg)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("[ERROR] unable to parse pkg %s due to error: %s\n", pkg, err))
		} else {
			fmt.Println(pkg)
			for _, req := range reqs {
				fmt.Printf("  %+v\n", req.Name)
			}
		}
	}
}
