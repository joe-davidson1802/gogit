package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/joerdav/gogit/cfg"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/oauth2"
)

func main() {
	cfg, err := cfg.LoadArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	client := newGitClient(cfg.GitPat)
	repos, err := client.getRepos(cfg.Org)
	if err != nil {
		panic(err)
	}
	ds := make(chan Diff, len(repos))
	wd := writeDiff
	for _, r := range repos {
		go func(r string) {
			diff, _ := client.getDiff(context.Background(), cfg.Org, r, cfg.Branch, cfg.Dep)
			ds <- diff
		}(r)
	}
	var diffs []Diff
	for range repos {
		d := <-ds
		if !semver.IsValid(d.Version) {
			continue
		}
		diffs = append(diffs, d)
	}
	sort.Slice(diffs, func(i, j int) bool {
		return semver.Compare(diffs[i].Version, diffs[j].Version) < 1
	})
	if cfg.Json {
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(diffs)
		fmt.Print(b.String())
		return
	}
	for _, d := range diffs {
		fmt.Print(wd(d.Name, d, cfg))
	}
}

func writeDiff(repoName string, diff Diff, cfg cfg.Config) string {
	if diff.Version == "" {
		return ""
	}
	sb := new(strings.Builder)
	msg := []string{}
	msg = append(msg, diff.Version)
	sb.WriteString(writeRow(repoName, strings.Join(msg, ", ")))
	return sb.String()
}

func writeRow(repoName, message string) string {
	return fmt.Sprintln(repoName, strings.Repeat(" ", 45-len(repoName)), message)
}

type gitClient struct {
	client *github.Client
}

func newGitClient(token string) gitClient {
	if token == "" {
		client := github.NewClient(nil)
		return gitClient{client}
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return gitClient{client}
}

func (cli gitClient) getRepos(org string) ([]string, error) {
	var names []string
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	// get all pages of results
	for {
		repos, resp, err := cli.client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			return names, err
		}
		for _, r := range repos {
			names = append(names, r.GetName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return names, nil
}

type Diff struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (cli gitClient) getDiff(ctx context.Context, org, repo, branch, packagepath string) (diff Diff, err error) {
	m, _, err := cli.client.Repositories.DownloadContents(ctx, org, repo, "go.mod", &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return
	}
	defer m.Close()
	b, err := ioutil.ReadAll(m)
	if err != nil {
		return
	}
	f, err := modfile.Parse("go.mod", b, nil)
	if err != nil {
		return
	}
	diff.Name = repo
	for _, r := range f.Require {
		if r.Mod.Path == packagepath {
			diff.Version = r.Mod.Version
			return
		}
	}
	diff.Version = "none"
	return
}
