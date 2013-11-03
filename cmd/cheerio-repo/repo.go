package main

import (
	"flag"
	"fmt"
	"github.com/beyang/cheerio/cheerio"
)

func main() {
	flag.Parse()
	pkg := cheerio.NormalizedPkgName(flag.Arg(0))

	repo, err := cheerio.DefaultPyPI.FetchSourceRepoURI(pkg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Println(repo)
	}
}
