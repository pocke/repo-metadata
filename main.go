package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/kylelemons/godebug/pretty"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v2"
)

type Option struct {
	RepoOwner   string
	RepoName    string
	AccessToken string
	DryRun      bool
}

type Configuration struct {
	Description *string
	Homepage    *string
	Topics      []string
}

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func Main(args []string) error {
	conf, err := loadConfiguration()
	if err != nil {
		return err
	}
	opt, err := parseOption(args)
	if err != nil {
		return err
	}

	if opt.DryRun {
		return dryRun(opt, conf)
	} else {
		return apply(opt, conf)
	}
}

func dryRun(opt *Option, conf *Configuration) error {
	ctx := context.Background()
	c := ghClient(ctx, opt.AccessToken)
	repo, _, err := c.Repositories.Get(ctx, opt.RepoOwner, opt.RepoName)
	if err != nil {
		return err
	}
	topics, _, err := c.Repositories.ListAllTopics(ctx, opt.RepoOwner, opt.RepoName)
	if err != nil {
		return err
	}

	cur := &Configuration{
		Description: repo.Description,
		Homepage:    repo.Homepage,
		Topics:      topics,
	}
	for _, line := range strings.Split(pretty.Compare(cur, conf), "\n") {
		if strings.HasPrefix(line, "+") {
			fmt.Print("\x1B[32m")
		} else if strings.HasPrefix(line, "-") {
			fmt.Print("\x1B[31m")
		}
		fmt.Println(line)
		fmt.Print("\x1B[0m")
	}
	return nil
}

func apply(opt *Option, conf *Configuration) error {
	ctx := context.Background()
	c := ghClient(ctx, opt.AccessToken)
	repo := &github.Repository{
		Name:        &opt.RepoName,
		Description: conf.Description,
		Homepage:    conf.Homepage,
	}
	_, _, err := c.Repositories.Edit(ctx, opt.RepoOwner, opt.RepoName, repo)
	if err != nil {
		return err
	}
	_, _, err = c.Repositories.ReplaceAllTopics(ctx, opt.RepoOwner, opt.RepoName, conf.Topics)
	if err != nil {
		return err
	}

	return nil
}

func loadConfiguration() (*Configuration, error) {
	path := "./.repo-metadata.yaml"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Configuration{}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func parseOption(args []string) (*Option, error) {
	opt := &Option{
		AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
	}
	if isTravis() {
		slug := strings.Split(os.Getenv("TRAVIS_REPO_SLUG"), "/")
		opt.RepoOwner = slug[0]
		opt.RepoName = slug[1]
		opt.DryRun = os.Getenv("TRAVIS_PULL_REQUEST") == "false"
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.StringVar(&opt.RepoOwner, "owner", opt.RepoOwner, "repository owner")
	fs.StringVar(&opt.RepoName, "name", opt.RepoName, "repository name")
	fs.StringVar(&opt.AccessToken, "access-token", opt.AccessToken, "GitHub access token")
	fs.BoolVar(&opt.DryRun, "dry-run", opt.DryRun, "dry run")
	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	return opt, nil
}

func ghClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func isTravis() bool {
	return os.Getenv("TRAVIS") == "true"
}

func lineDiff(a, b string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	return dmp.DiffMain(a, b, true)
}
