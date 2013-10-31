package main

import (
	"flag"
	"fmt"
	ppg "github.com/beyang/pypigraph"
	ppq "github.com/beyang/pypigraph/pypiquery"
	"os"
	"strings"
)

var file = flag.String("graphfile", "", "Path to graph file.  Defaults to $GOPATH/src/github.com/beyang/pypigraph/data/pypi_graph")

func main() {
	flag.Parse()
	pkg := ppg.NormalizedPkgName(flag.Arg(0))

	var pypi *ppq.PyPIGraph
	if *file == "" {
		pypi = ppq.DefaultPyPI
	} else {
		var err error
		pypi, err = ppq.NewPyPIGraph(*file)
		if err != nil {
			fmt.Printf("Error creating PyPI graph: %s\n", err)
			os.Exit(1)
		}
	}

	pkgReq := pypi.Requires(pkg)
	pkgReqBy := pypi.RequiredBy(pkg)
	fmt.Printf("pkg %s uses (%d):\n  %s\nand is used by (%d):\n  %s\n", pkg, len(pkgReq), strings.Join(pkgReq, " "), len(pkgReqBy), strings.Join(pkgReqBy, " "))
}
