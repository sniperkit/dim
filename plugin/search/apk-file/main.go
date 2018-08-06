package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	// external
	"github.com/PuerkitoBio/goquery"
	gotl "github.com/agrison/go-tablib"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"

	// internal
	"github.com/sniperkit/dim/plugin/search/apk-file/version"
	// dim "github.com/sniperkit/dim/pkg/core"
	// "github.com/sniperkit/dim/pkg/config"
	// "github.com/sniperkit/dim/pkg/plugin"
)

const (
	pluginName              = "apk-file"
	alpineContentsSearchURI = "https://pkgs.alpinelinux.org/contents"
)

type fileInfo struct {
	path, pkg, branch, repo, arch string
}

type FileInfo struct {
	Path    string `json:"path" yaml:"path" toml:"path" ini:"path" xml:"path" csv:"Path"`
	Package string `json:"package" yaml:"package" toml:"package" ini:"package" xml:"package" csv:"Package"`
	Branch  string `json:"branch" yaml:"branch" toml:"branch" ini:"branch" xml:"branch" csv:"Branch"`
	Repo    string `json:"repo" yaml:"repo" toml:"repo" ini:"repo" xml:"repo" csv:"Repository"`
	Arch    string `json:"arch" yaml:"arch" toml:"arch" ini:"arch" xml:"arch" csv:"Architecture"`
}

var (
	// search args
	arch, branch, repo string

	// export args
	prefixPath, basename, format string

	// dev args
	debug, verbose, vrsn, save bool

	// argument validation lists
	validBranches = []string{"edge", "v3.8", "v3.7", "v3.6", "v3.5", "v3.4", "v3.3"}
	validArches   = []string{"x86", "x86_64", "armhf"}
	validRepos    = []string{"main", "community", "testing"}
	validOutput   = []string{"markdown", "csv", "yaml", "json", "xlsx", "xml", "tsv", "mysql", "postgres", "html", "ascii"}
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "apk-file"
	p.Description = "Search apk package contents via the command line"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&arch, "arch", "x86_64", "arch to search for ("+strings.Join(validArches, ", ")+")")
	p.FlagSet.StringVar(&branch, "branch", "v3.8", "repository to search in ("+strings.Join(validBranches, ", ")+")")
	p.FlagSet.StringVar(&repo, "repo", "", "repository to search in ("+strings.Join(validRepos, ", ")+")")
	p.FlagSet.StringVar(&format, "format", "", "format results with  ("+strings.Join(validOutput, ", ")+") format.")
	p.FlagSet.StringVar(&prefixPath, "output", "./output", "output results to prefix_path (default: ./output).")
	p.FlagSet.StringVar(&basename, "basename", "results", "output results to file with basename: (default: ./basename.[FORMAT]).")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&verbose, "v", true, "enable verbose mode")

	// p.FlagSet.BoolVar(&vrsn, "version", false, "print version and exit")
	// p.FlagSet.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if arch != "" && !stringInSlice(arch, validArches) {
			return fmt.Errorf("%s is not a valid arch", arch)
		}

		if repo != "" && !stringInSlice(repo, validRepos) {
			return fmt.Errorf("%s is not a valid repo", repo)
		}

		if branch != "" && !stringInSlice(branch, validBranches) {
			return fmt.Errorf("%s is not a valid branch", branch)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {

		if len(args) < 1 {
			// if p.FlagSet.NArg() < 1 {
			return errors.New("must pass a file to search for")
		}

		if format == "" {
			format = "yaml"
		}

		if _, err := os.Stat(prefixPath); os.IsNotExist(err) {
			os.Mkdir(prefixPath, 0700)
		}

		outputPrefixBasePath := prefixPath + "/" + basename

		queryStr := p.FlagSet.Arg(0)

		f, p := getFileAndPath(p.FlagSet.Arg(0))

		query := url.Values{
			"file":   {f},
			"path":   {p},
			"branch": {branch},
			"repo":   {repo},
			"arch":   {arch},
		}

		uri := fmt.Sprintf("%s?%s", alpineContentsSearchURI, query.Encode())
		logrus.Debugf("requesting from %s", uri)
		resp, err := http.Get(uri)
		if err != nil {
			logrus.Fatalf("requesting %s failed: %v", uri, err)
		}
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			logrus.Fatalf("creating document failed: %v", err)
		}

		// w.Flush()
		if err := exportResults(queryStr, getFilesInfo(doc), outputPrefixBasePath, format, verbose); err != nil {
			return err
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func exportResults(query string, files []fileInfo, outputFile string, outputFormat string, verbose bool) error {

	ds := gotl.NewDataset([]string{"query", "search_at", "file", "package", "branch", "repository", "architecture"})
	for _, f := range files {
		ds.AppendValues(query, time.Now(), f.path, f.pkg, f.branch, f.repo, f.arch)
	}

	outputFile = outputFile + "." + outputFormat

	var err error
	var outputTab *gotl.Exportable

	switch outputFormat {
	case "csv":
		outputTab, err = ds.CSV()
	case "tsv":
		outputTab, err = ds.TSV()
	case "yaml", "yml":
		outputTab, err = ds.YAML()
	case "json":
		outputTab, err = ds.JSON()
	case "xlsx":
		outputTab, err = ds.XLSX()
	case "xml":
		outputTab, err = ds.XML()
	case "mysql":
		// todo: add search column, add search date
		outputTab = ds.MySQL(pluginName)
	case "postgres":
		// todo: add search column, add search date
		outputTab = ds.Postgres(pluginName)
	case "html":
		outputTab = ds.HTML()
	case "ascii":
		fallthrough
	default:
		outputTab = ds.Tabular("condensed" /* tablib.TabularGrid */)
	}
	if err != nil {
		return err
	}

	if err = outputTab.WriteFile(outputFile, 0644); err != nil {
		return err
	}

	if verbose {
		outputTab = ds.Tabular("condensed" /* tablib.TabularGrid */)
		fmt.Println(outputTab)
	}

	return nil
}

func getFilesInfo(d *goquery.Document) []fileInfo {
	files := []fileInfo{}
	d.Find(".pure-table tr:not(:first-child)").Each(func(j int, l *goquery.Selection) {
		f := fileInfo{}
		rows := l.Find("td")

		// pp.Println(rows)

		rows.Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				f.path = s.Text()
			case 1:
				f.pkg = s.Text()
			case 2:
				f.branch = s.Text()
			case 3:
				f.repo = s.Text()
			case 4:
				f.arch = s.Text()
			default:
				logrus.Warn("Unmapped value for column %d with value %s", i, s.Text())
			}
		})
		files = append(files, f)
	})
	return files
}

func getFileAndPath(arg string) (file string, dir string) {
	file = "*" + path.Base(arg) + "*"
	dir = path.Dir(arg)
	if dir != "" && dir != "." {
		dir = "*" + dir
		file = strings.TrimPrefix(file, "*")
	} else {
		dir = ""
	}
	return file, dir
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
