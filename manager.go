package main

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"os"
    "flag"
)

func main() {

	client, _ := getGithubClient()
	owner := "Ariba"
	repos, _ := getOrgRepos(client, owner)

	for _, repo := range repos {
		pulls, _ := getRepoPulls(client, owner, *repo.Name)
		for _, pull := range pulls {
			log.Println(*pull.ID, *pull.Number, *pull.Title)
		}
	}

	log.Println("Main complete")
}

func getGithubClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	baseUrl := os.Getenv("GITHUB_API_URL")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewEnterpriseClient(baseUrl, baseUrl, tc)
}
func getOrgRepos(client *github.Client, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{}
	for {
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}

		opt.Page = resp.NextPage
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
	}
	return allRepos, nil
}

func getRepoPulls(client *github.Client, owner string, repo string) ([]*github.PullRequest, error) {
	var allPulls []*github.PullRequest
	opt := &github.PullRequestListOptions{}
	for {
		pulls, resp, err := client.PullRequests.List(context.Background(), owner, repo, opt)
		if err != nil {
			return nil, err
		}
		if len(pulls) == 0 {
			break
		}

		opt.Page = resp.NextPage
		allPulls = append(allPulls, pulls...)
		if resp.NextPage == 0 {
			break
		}
	}
	return allPulls, nil
}
