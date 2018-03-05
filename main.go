package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/vcs"
)

func main() {
	downloadAndBuild("gotools", tools{
		"golang.org/x/lint/golint": {
			"golang.org/x/lint":  "",
			"golang.org/x/tools": "",
		},
		"golang.org/x/tools/cmd/gorename": {
			"golang.org/x/tools": "",
		},
		"golang.org/x/tools/cmd/guru": {
			"golang.org/x/tools": "",
		},
		"golang.org/x/tools/cmd/goimports": {
			"golang.org/x/tools": "",
		},
		"golang.org/x/tools/cmd/gotype": {
			"golang.org/x/tools": "",
		},
		"golang.org/x/tools/cmd/eg": {
			"golang.org/x/tools": "",
		},
		"github.com/derekparker/delve/cmd/dlv": {
			"github.com/derekparker/delve": "",
		},
		"honnef.co/go/tools/cmd/megacheck": {
			"honnef.co/go/tools":        "",
			"golang.org/x/tools":        "",
			"github.com/kisielk/gotool": "",
		},
		"github.com/shurcooL/binstale": {
			"github.com/shurcooL/binstale": "",
		},
		"github.com/shurcooL/Go-Package-Store": {
			"github.com/shurcooL/Go-Package-Store": "",
			"golang.org/x/tools":                   "",
			"github.com/shurcooL/vcsstate":         "",
			"github.com/shurcooL/go/trim":          "",
			"github.com/kisielk/gotool":            "",
			"github.com/bradfitz/iter":             "",
		},
		"github.com/shurcooL/gostatus": {
			"github.com/shurcooL/gostatus": "",
		},
		"github.com/rjeczalik/bin/cmd/gobin": {
			"github.com/rjeczalik/bin":   "",
			"github.com/rjeczalik/which": "",
		},
		"github.com/rogpeppe/govers": {
			"github.com/rogpeppe/govers": "",
		},
		"github.com/loov/view-annotated-file": {
			"github.com/loov/view-annotated-file": "",
		},
		"github.com/awalterschulze/goderive": {
			"github.com/awalterschulze/goderive": "",
		},
		//"github.com/zyedidia/micro/cmd/micro": {
		//	"github.com/blang/semver":                 "",
		//	"github.com/dustin/go-humanize":           "",
		//	"github.com/flynn/json5":                  "",
		//	"github.com/gdamore/encoding":             "",
		//	"github.com/go-errors/errors":             "",
		//	"github.com/kr/pty":                       "",
		//	"github.com/lucasb-eyer/go-colorful":      "",
		//	"github.com/mattn/go-isatty":              "",
		//	"github.com/mattn/go-runewidth":           "",
		//	"github.com/mitchellh/go-homedir":         "",
		//	"github.com/sergi/go-diff/diffmatchpatch": "",
		//	"github.com/yuin/gopher-lua":              "",
		//	"github.com/zyedidia/clipboard":           "",
		//	"github.com/zyedidia/glob":                "",
		//	"github.com/zyedidia/micro":               "",
		//	"github.com/zyedidia/tcell":               "",
		//	"github.com/zyedidia/terminal":            "",
		//	"github.com/zyedidia/pty":                 "",
		//	"golang.org/x/text":                       "",
		//	"gopkg.in/yaml.v2":                        "",
		//	"layeh.com/gopher-luar":                   "",
		//},
	})

	//err := parrallelBuild(
	//	"github.com/KyleBanks/depth/cmd/depth",
	//	"github.com/haya14busa/goverage",

	//	"github.com/sqs/goreturns",
	//	"github.com/lukehoban/go-outline",
	//	"github.com/newhook/go-symbols",
	//	"github.com/nsf/gocode",
	//	"github.com/rogpeppe/godef",
	//	"github.com/tpng/gopkgs",

	//	"github.com/mibk/dupl",
	//	"github.com/kisielk/errcheck",
	//	"github.com/fzipp/gocyclo",
	//	"mvdan.cc/interfacer",
	//	"github.com/mdempsky/unconvert",
	//)
}

type tools map[string]map[string]string

func downloadAndBuild(folder string, t tools) error {
	toolsDir, err := GetToolsDir()
	if err != nil {
		return err
	}

	defer os.RemoveAll(toolsDir)
	os.Setenv("GOPATH", toolsDir)

	var totalDeps []depInfo
	for _, deps := range t {
		for repo, rev := range deps {
			newDep := depInfo{
				repo: repo,
				rev:  rev,
			}

			if !deriveContains(totalDeps, newDep) {
				totalDeps = append(totalDeps, newDep)
			}
		}
	}

	err = downloadDepsTo(toolsDir, totalDeps)
	if err != nil {
		return err
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	binariesDir := filepath.Join(usr.HomeDir, folder)
	err = os.MkdirAll(binariesDir, 0700)
	if err != nil {
		return err
	}

	err = os.Chdir(binariesDir)
	if err != nil {
		return err
	}

	var toBuild []string
	for tool := range t {
		toBuild = append(toBuild, tool)
	}

	return parrallelBuild(toBuild...)
}

type depInfo struct {
	repo string
	rev  string
}

func downloadDepsTo(folder string, deps []depInfo) error {
	runCount := runtime.GOMAXPROCS(0) * 2
	var wg sync.WaitGroup
	wg.Add(runCount)

	depsCount := len(deps)
	errorCh := make(chan error, depsCount)
	go func() {
		wg.Wait()
		close(errorCh)
	}()

	parts := SplitIntoParts(depsCount, runCount)
	for _, part := range parts {
		start := part.Begin
		end := part.End
		go func() {
			defer wg.Done()
			deps := deps[start:end]
			for _, dep := range deps {
				log.Println("Downloading", dep.repo)
				_, err := Download(dep.repo, "", filepath.Join(folder, "src"), dep.rev)
				log.Println("Done with", dep.repo)
				errorCh <- errors.WithMessage(err, "failed to get "+dep.repo)
			}
		}()
	}

	for err := range errorCh {
		if err != nil {
			return err
		}
	}

	return nil
}

type VCS struct {
	Root       string
	ImportPath string
	Type       string
}

func Download(importPath, repoPath, target, rev string) (*VCS, error) {
	// Analyze the import path to determine the version control system,
	// repository, and the import path for the root of the repository.
	rr, err := vcs.RepoRootForImportPath(importPath, false)
	if err != nil {
		return nil, err
	}

	root := filepath.Join(target, rr.Root)
	if repoPath != "" {
		rr.Repo = repoPath
	}

	if err := os.RemoveAll(root); err != nil {
		return nil, fmt.Errorf("remove package root: %v", err)
	}
	// Some version control tools require the parent of the target to exist.
	parent, _ := filepath.Split(root)
	if err = os.MkdirAll(parent, 0777); err != nil {
		return nil, err
	}

	if rev == "" {
		if err = rr.VCS.Create(root, rr.Repo); err != nil {
			return nil, err
		}
	} else {
		if err = rr.VCS.CreateAtRev(root, rr.Repo, rev); err != nil {
			return nil, err
		}
	}
	return &VCS{Root: root, ImportPath: rr.Root, Type: rr.VCS.Cmd}, nil
}

func parrallelBuild(paths ...string) error {
	cpunum := runtime.NumCPU()
	parts := SplitIntoParts(len(paths), cpunum)

	group, ctx := errgroup.WithContext(context.Background())
	for _, part := range parts {
		toBuildSlice := paths[part.Begin:part.End]
		group.Go(func() error {
			for _, toBuild := range toBuildSlice {
				log.Println("Working with", toBuild)

				cmd := exec.CommandContext(ctx, "go", "build", toBuild)
				_, err := cmd.Output()
				log.Println("Finished with", toBuild)
				if err != nil {
					switch v := err.(type) {
					case *exec.ExitError:
						log.Println("Error with", toBuild, "error is", string(v.Stderr))
					default:
						log.Println("Unknown Error with", toBuild, "error is", v)
					}
					return err
				}
			}
			return nil
		})
	}

	return group.Wait()
}

type Part struct {
	Begin int
	End   int
}

// SplitIntoParts returns at MOST N Parts for even distribution for passed length
// It's used so we can divide our big log slice between goroutines
func SplitIntoParts(length, N int) []Part {
	chunkSize := length / N
	bonus := length - chunkSize*N // i.e. remainder

	var parts []Part
	start, end := 0, chunkSize
	for start < length {
		if bonus > 0 {
			end++
			bonus--
		}
		/* do something with array slice over [start, end) interval */
		parts = append(parts, Part{start, end})

		start = end
		end = start + chunkSize
	}

	return parts
}

func GetToolsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	toolsdir := filepath.Join(wd, "toolsdir")
	err = os.MkdirAll(toolsdir, 0700)
	if err != nil {
		return "", err
	}

	return toolsdir, nil
}

// deriveContains returns whether the item is contained in the list.
func deriveContains(list []depInfo, item depInfo) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
