package main

import (
	"flag"
	"fmt"
	"github.com/beyang/cheerio"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	Cmd_Repo     = "repo"
	Cmd_Reqs     = "reqs"
	Cmd_ReqGen   = "reqs-generate"
	Cmd_TopLevel = "toplevel"
)

var Commands = map[string]func(args []string, flags *flag.FlagSet){
	Cmd_Repo:     mainRepo,
	Cmd_Reqs:     mainReqs,
	Cmd_ReqGen:   mainReqGen,
	Cmd_TopLevel: mainTopLevel,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [command-opts]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Commands:")
		for cmd, _ := range Commands {
			fmt.Fprintf(os.Stderr, "  %s\n", cmd)
		}
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	subcommand := flag.Arg(0)

	if cmd, in := Commands[subcommand]; in {
		flags := flag.NewFlagSet(Cmd_Repo, flag.ExitOnError)
		cmd(os.Args[1:], flags)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "Subcommand not recognized: %s\n", flag.Arg(0))
	flag.Usage()
	os.Exit(1)
}

func mainRepo(args []string, flags *flag.FlagSet) {
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s <package-name>\n", os.Args[0], args[0])
	}
	flags.Parse(args[1:])

	if flags.NArg() < 1 {
		flags.Usage()
		os.Exit(1)
	}

	pkg := cheerio.NormalizedPkgName(flags.Arg(0))

	repo, err := cheerio.DefaultPyPI.FetchSourceRepoURI(pkg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Println(repo)
	}
}

func mainTopLevel(args []string, flags *flag.FlagSet) {
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s <package-name>\n", os.Args[0], args[0])
	}
	flags.Parse(args[1:])

	if flags.NArg() < 1 {
		flags.Usage()
		os.Exit(1)
	}

	pkg := cheerio.NormalizedPkgName(flags.Arg(0))

	modules, err := cheerio.DefaultPyPI.FetchSourceTopLevelModules(pkg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Println(strings.Join(modules, " "))
	}
}

func mainReqs(args []string, flags *flag.FlagSet) {
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s <package-name>\n", os.Args[0], args[0])
		flags.PrintDefaults()
	}
	file := flags.String("graphfile", "", fmt.Sprintf("Path to PyPI dependency graph file.  Defaults to $GOPATH/src/github.com/beyang/cheerio/data/pypi_graph"))
	flags.Parse(args[1:])

	if flags.NArg() < 1 {
		flags.Usage()
		os.Exit(1)
	}

	pkg := cheerio.NormalizedPkgName(flags.Arg(0))

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

// Prints PyPI requirement graph to stdout in the below format. Skips errors (including packages where there is no requires.txt file).
// Example format:
//
// pkg1
// pkg1:pkg2
// pkg1:pkg3
// pkg2
// pkg2:pkg4
func mainReqGen(args []string, flags *flag.FlagSet) {
	pkgIndex := &cheerio.DefaultPyPI
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

			reqs, err := pkgIndex.FetchPackageRequirements(pkg)
			if err != nil {
				if !strings.Contains(err.Error(), "No file matched pattern") { // ignore archives that don't contain requires.txt
					os.Stderr.WriteString(fmt.Sprintf("[ERROR] unable to parse pkg %s due to error: %s\n", pkg, err))
				}
			} else {
				stdoutMu.Lock()
				fmt.Println(cheerio.NormalizedPkgName(pkg))
				for _, req := range reqs {
					fmt.Printf("%s:%s\n", cheerio.NormalizedPkgName(pkg), cheerio.NormalizedPkgName(req.Name))
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
