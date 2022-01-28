package cfg

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const defaultPath = "~/.gogitpat"

type cmdArgs struct {
	pat, patPath     string
	org, branch, dep string
	help, json       bool
}

type Config struct {
	GitPat           string
	Org, Branch, Dep string
	Json             bool
}

func LoadArgs() (cfg Config, err error) {
	var cmd cmdArgs

	flag.Usage = func() {
		fmt.Println("usage: gogit [owner] [branch] [mod] [-pat | -pat-path] ")
		flag.PrintDefaults()
	}
	flag.StringVar(&cmd.patPath, "pat-path", "", "A file containing your github PAT.")
	flag.StringVar(&cmd.pat, "pat", "", "Your github PAT.")
	flag.BoolVar(&cmd.help, "h", false, "Show help text.")
	flag.BoolVar(&cmd.json, "json", false, "Output as json.")

	flag.Parse()

	args := flag.Args()
	if cmd.pat == "" && cmd.patPath == "" {
		err = errors.New("Either pat or pat-path must be provided")
		flag.Usage()
		return
	}
	if cmd.help {
		flag.Usage()
		return
	}
	cfg = Config{
		GitPat: cmd.pat,
		Org:    args[0],
		Branch: args[1],
		Dep:    args[2],
		Json:   cmd.json,
	}
	if cfg.GitPat == "" {
		cfg.GitPat = loadFromPath(cmd.patPath)
	}
	return
}

func loadFromPath(p string) string {
	path := defaultPath
	if path != "" {
		path = p
	}
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}
	b, _ := os.ReadFile(path)
	return strings.TrimSpace(string(b))
}
