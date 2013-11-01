package req

import (
	"bufio"
	"fmt"
	"github.com/beyang/cheerio/py"
	"os"
	"path/filepath"
	"strings"
)

var DefaultPyPI *PyPIGraph
var DefaultPyPIFile = filepath.Join(os.Getenv("GOPATH"), "src/github.com/beyang/cheerio/data/pypi_graph")

func init() {
	var err error
	DefaultPyPI, err = NewPyPIGraph(DefaultPyPIFile)
	if err != nil {
		panic(fmt.Sprintf("Cannot initialize default PyPI because: %s", err))
	}
}

type PyPIGraph struct {
	Req   map[string][]string
	ReqBy map[string][]string
}

func NewPyPIGraph(file string) (*PyPIGraph, error) {
	var graph *PyPIGraph

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	graph = &PyPIGraph{
		Req:   make(map[string][]string),
		ReqBy: make(map[string][]string),
	}
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

				if _, in := graph.Req[pkg]; !in {
					graph.Req[pkg] = make([]string, 0)
				}
				graph.Req[pkg] = append(graph.Req[pkg], dep)

				if _, in := graph.ReqBy[dep]; !in {
					graph.ReqBy[dep] = make([]string, 0)
				}
				graph.ReqBy[dep] = append(graph.ReqBy[dep], pkg)
			}
		} else if line != "" {
			pkg := line
			if _, in := graph.Req[pkg]; !in {
				graph.Req[pkg] = make([]string, 0)
			}
			if _, in := graph.ReqBy[pkg]; !in {
				graph.ReqBy[pkg] = make([]string, 0)
			}
		}
	}

	return graph, nil
}

func (p *PyPIGraph) Requires(pkg string) []string {
	return p.Req[py.NormalizedPkgName(pkg)]
}

func (p *PyPIGraph) RequiredBy(pkg string) []string {
	return p.ReqBy[py.NormalizedPkgName(pkg)]
}
