package main

import (
	"fmt"
	ppg "github.com/beyang/pypigraph"
	"log"
	"os"
	"sync"
)

func main() {
	pkgIndex := &ppg.PackageIndex{URI: "https://pypi.python.org"}
	pkgs, err := pkgIndex.AllPackages()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[FATAL] %s\n", err))
		os.Exit(1)
	}

	var stdoutMu sync.Mutex
	var pkgsCompleteMu sync.Mutex
	var waiter sync.WaitGroup
	throttle := make(chan int, 100)
	pkgsComplete := 0
	for p, pkg_ := range pkgs {
		pkg := pkg_

		waiter.Add(1)
		throttle <- p
		go func() {
			defer waiter.Done()
			defer func() { <-throttle }()

			reqs, err := pkgIndex.PackageRequirements(pkg)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("[ERROR] unable to parse pkg %s due to error: %s\n", pkg, err))
			} else {
				stdoutMu.Lock()
				fmt.Println(ppg.NormalizedPkgName(pkg))
				for _, req := range reqs {
					fmt.Printf("%s:%s\n", ppg.NormalizedPkgName(pkg), ppg.NormalizedPkgName(req.Name))
				}
				stdoutMu.Unlock()
			}

			pkgsCompleteMu.Lock()
			if pkgsComplete%50 == 0 {
				log.Printf("[status] %d / %d\n", pkgsComplete, len(pkgs))
			}
			pkgsComplete++
			pkgsCompleteMu.Unlock()
		}()
	}
	waiter.Wait()
}
