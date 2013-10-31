package main

import (
	"fmt"
	ppg "github.com/beyang/pypigraph"
	"log"
	"os"
)

func main() {
	pkgIndex := &ppg.PackageIndex{URI: "https://pypi.python.org"}
	pkgs, err := pkgIndex.AllPackages()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[FATAL] %s\n", err))
		os.Exit(1)
	}

	for p, pkg := range pkgs {
		if p%50 == 0 { // progress
			log.Printf("[status] %d / %d\n", p, len(pkgs))
		}

		reqs, err := pkgIndex.PackageRequirements(pkg)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("[ERROR] unable to parse pkg %s due to error: %s\n", pkg, err))
		} else {
			for _, req := range reqs {
				fmt.Printf("%s:%s\n", ppg.NormalizedPkgName(pkg), ppg.NormalizedPkgName(req.Name))
			}
		}
	}
}
