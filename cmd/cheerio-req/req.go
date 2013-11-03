package main

import (
	"flag"
	"fmt"
	"github.com/beyang/cheerio/cheerio"
	"os"
	"strings"
)

var file = flag.String("graphfile", "", fmt.Sprintf("Path to PyPI dependency graph file.  Defaults to %s", cheerio.DefaultPyPIGraphFile))

func main() {
	flag.Parse()
	pkg := cheerio.NormalizedPkgName(flag.Arg(0))

	var pypiG *cheerio.PyPIGraph
	if *file == "" {
		pypiG = cheerio.DefaultPyPIGraph
	} else {
		var err error
		pypiG, err = cheerio.NewPyPIGraph(*file)
		if err != nil {
			fmt.Printf("Error creating PyPI graph: %s\n", err)
			os.Exit(1)
		}
	}

	pkgReq := pypiG.Requires(pkg)
	pkgReqBy := pypiG.RequiredBy(pkg)
	fmt.Printf("pkg %s uses (%d):\n  %s\nand is used by (%d):\n  %s\n", pkg, len(pkgReq), strings.Join(pkgReq, " "), len(pkgReqBy), strings.Join(pkgReqBy, " "))
}
