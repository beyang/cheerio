package main

import (
	"bufio"
	"flag"
	"fmt"
	ppg "github.com/beyang/pypigraph"
	"os"
	"path/filepath"
	"strings"
)

var file = flag.String("graphfile", "", "Path to graph file.  Defaults to $GOPATH/src/github.com/beyang/pypigraph/data/pypi_graph")

func main() {
	flag.Parse()
	pkg := ppg.NormalizedPkgName(flag.Arg(0))

	if *file == "" {
		*file = filepath.Join(os.Getenv("GOPATH"), "src/github.com/beyang/pypigraph/data/pypi_graph")
	}

	requires := make(map[string][]string)
	requiredBy := make(map[string][]string)

	f, _ := os.Open(*file)
	reader := bufio.NewReader(f)
	for {
		lineB, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		line := string(lineB)

		if strings.Contains(line, ":") {
			lineSplit := strings.Split(line, ":")
			if len(lineSplit) == 2 {
				pkg, dep := lineSplit[0], lineSplit[1]

				if _, in := requires[pkg]; !in {
					requires[pkg] = make([]string, 0)
				}
				requires[pkg] = append(requires[pkg], dep)

				if _, in := requiredBy[dep]; !in {
					requiredBy[dep] = make([]string, 0)
				}
				requiredBy[dep] = append(requiredBy[dep], pkg)
			}
		} else if line != "" {
			if _, in := requires[pkg]; !in {
				requires[pkg] = make([]string, 0)
			}
			if _, in := requiredBy[pkg]; !in {
				requiredBy[pkg] = make([]string, 0)
			}
		}
	}

	uses := requires[pkg]
	usedBy := requiredBy[pkg]
	fmt.Printf("pkg %s uses (%d):\n  %s\nand is used by (%d):\n  %s\n", pkg, len(uses), strings.Join(uses, " "), len(usedBy), strings.Join(usedBy, " "))
}
