package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

var allPkgRegexp = regexp.MustCompile(``)

func allPkgs(uri string) ([]string, error) {
	pkgs := make([]string, 0)

	resp, err := http.Client.Get(uri)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	matches := allPkgRegexp.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		if len(match) != 3 {
			return fmt.Errorf("Unexpected number of matches")
		} else if match[1] != match[2] {
			return fmt.Errorf("Names do not match %s != %s", match[1], match[2])
		} else {
			pkgs = append(pkgs, match[1])
		}
	}

	return pkgs
}

func main() {
	pypiURI := "http://pypi.python.org"
	pkgs := allPkgs(pypiURI)
	for _, pkg := range pkgs {
		fmt.Println(pkg)
	}
}
