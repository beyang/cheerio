package main

import (
	"flag"
	"fmt"
	"github.com/beyang/cheerio/py"
	"github.com/beyang/cheerio/req"
	"os"
	"strings"
)

var file = flag.String("graphfile", "", fmt.Sprintf("Path to PyPI dependency graph file.  Defaults to %s", req.DefaultPyPIFile))

func main() {
	flag.Parse()
	pkg := py.NormalizedPkgName(flag.Arg(0))

	var pypiG *req.PyPIGraph
	if *file == "" {
		pypiG = req.DefaultPyPI
	} else {
		var err error
		pypiG, err = req.NewPyPIGraph(*file)
		if err != nil {
			fmt.Printf("Error creating PyPI graph: %s\n", err)
			os.Exit(1)
		}
	}

	pkgReq := pypiG.Requires(pkg)
	pkgReqBy := pypiG.RequiredBy(pkg)
	fmt.Printf("pkg %s uses (%d):\n  %s\nand is used by (%d):\n  %s\n", pkg, len(pkgReq), strings.Join(pkgReq, " "), len(pkgReqBy), strings.Join(pkgReqBy, " "))
}
