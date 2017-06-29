package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/LK4D4/vndr/godl"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func parrallelBuild(paths ...string) error {
	cpunum := runtime.NumCPU()
	parts := SplitIntoParts(len(paths), cpunum)

	group, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < cpunum; i++ {
		toBuildSlice := paths[parts[i].Begin:parts[i].End]
		group.Go(func() error {
			for _, toBuild := range toBuildSlice {
				log.Println("Working with", toBuild)

				cmd := exec.CommandContext(ctx, "go1.9beta2", "build", toBuild)
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

func MustNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	wd, err := os.Getwd()
	MustNoError(err)

	toolsdir := filepath.Join(wd, "toolsdir")
	MustNoError(os.MkdirAll(toolsdir, 0700))
	defer os.RemoveAll(toolsdir)

	os.Setenv("GOPATH", toolsdir)
	err = downloadDeps(toolsdir,
		"github.com/sqs/goreturns", "",
		"golang.org/x/tools", "",
		"github.com/golang/lint/golint", "",
		"github.com/lukehoban/go-outline", "",
		"github.com/newhook/go-symbols", "",
		"github.com/nsf/gocode", "",
		"github.com/rogpeppe/godef", "",
		"github.com/tpng/gopkgs", "",
		"github.com/derekparker/delve/cmd/dlv", "",
		"github.com/kisielk/gotool", "",
		"honnef.co/go/tools", "",

		"github.com/LK4D4/vndr", "",
		"github.com/shurcooL/binstale", "",
		"github.com/shurcooL/Go-Package-Store", "",
		"github.com/shurcooL/gostatus", "",
		"github.com/uber/go-torch", "",
		"github.com/rjeczalik/bin", "",
		"github.com/KyleBanks/depth", "",
		"github.com/haya14busa/goverage", "",
		"github.com/rogpeppe/govers", "",
		"github.com/zyedidia/micro/cmd/micro", "",

		"github.com/alecthomas/gometalinter", "",
		"github.com/opennota/check", "",
		"github.com/tsenart/deadcode", "",
		"github.com/mibk/dupl", "",
		"github.com/kisielk/errcheck", "",
		"github.com/GoASTScanner/gas", "",
		"github.com/jgautheron/goconst", "",
		"github.com/fzipp/gocyclo", "",
		"github.com/mvdan/interfacer", "",
		"github.com/gordonklaus/ineffassign", "",
		"github.com/mdempsky/unconvert", "",

		"github.com/mvdan/lint", "",
		"github.com/fatih/color", "",
		"github.com/rjeczalik/which", "",
		"github.com/shurcooL/vcsstate", "",
		"github.com/jessevdk/go-flags", "",
		"github.com/shurcooL/go", "",
		"github.com/bradfitz/iter", "",
		"golang.org/x/text", "",
	)
	MustNoError(err)

	usr, err := user.Current()
	MustNoError(err)

	binariesDir := filepath.Join(usr.HomeDir, "gotools")
	MustNoError(os.MkdirAll(binariesDir, 0700))
	MustNoError(os.Chdir(binariesDir))

	err = parrallelBuild(
		"github.com/sqs/goreturns",
		"golang.org/x/tools/cmd/gorename",
		"golang.org/x/tools/cmd/guru",
		"golang.org/x/tools/cmd/goimports",
		"golang.org/x/tools/cmd/gotype",
		"github.com/golang/lint/golint",
		"github.com/lukehoban/go-outline",
		"github.com/newhook/go-symbols",
		"github.com/nsf/gocode",
		"github.com/rogpeppe/godef",
		"github.com/tpng/gopkgs",
		"github.com/derekparker/delve/cmd/dlv",
		"honnef.co/go/tools/cmd/unused",
		"honnef.co/go/tools/cmd/gosimple",
		"honnef.co/go/tools/cmd/staticcheck",

		"github.com/LK4D4/vndr",
		"github.com/shurcooL/binstale",
		"github.com/shurcooL/Go-Package-Store",
		"github.com/shurcooL/gostatus",
		"github.com/uber/go-torch",
		"github.com/rjeczalik/bin/cmd/gobin",
		"github.com/KyleBanks/depth/cmd/depth",
		"github.com/haya14busa/goverage",
		"github.com/rogpeppe/govers",
		"github.com/rogpeppe/govers",
		"github.com/zyedidia/micro/cmd/micro",

		"github.com/alecthomas/gometalinter",
		"github.com/opennota/check/cmd/aligncheck",
		"github.com/opennota/check/cmd/structcheck",
		"github.com/opennota/check/cmd/varcheck",
		"github.com/tsenart/deadcode",
		"github.com/mibk/dupl",
		"github.com/kisielk/errcheck",
		"github.com/GoASTScanner/gas",
		"github.com/jgautheron/goconst",
		"github.com/fzipp/gocyclo",
		"github.com/mvdan/interfacer/cmd/interfacer",
		"github.com/gordonklaus/ineffassign",
		"github.com/mdempsky/unconvert",
	)
	MustNoError(err)
}

func downloadDeps(repoFolder string, deps ...string) error {
	if len(deps) == 0 || len(deps)%2 != 0 {
		return errors.New("Wrong arguments number")
	}

	var group errgroup.Group
	for i := 0; i < len(deps); i += 2 {
		dep := deps[i]
		ref := deps[i+1]
		group.Go(func() error {
			_, err := godl.Download(dep, "", filepath.Join(repoFolder, "src"), ref)
			if err != nil {
				return errors.WithMessage(err, "failed to get"+dep)
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
