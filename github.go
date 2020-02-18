package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getGithubClient(ag Arguments) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ag.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewEnterpriseClient(ag.apiURL, ag.apiURL, tc)
}
func getOrgRepos(client *github.Client, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{}
	for {
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			log.Fatal("Unable to fetch respositories for the organization: ", err)
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
			log.Fatal("Unable to fetch pull requests for the repository: ", err)
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

func getRepoBranches(client *github.Client, owner string, repo string) ([]*github.Branch, error) {
	var allBranches []*github.Branch
	opt := &github.BranchListOptions{}
	for {
		branches, resp, err := client.Repositories.ListBranches(context.Background(), owner, repo, opt)
		if err != nil {
			log.Fatal("Unable to fetch branches for the repository: ", err)
			return nil, err
		}
		if len(branches) == 0 {
			break
		}

		opt.Page = resp.NextPage
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
	}
	return allBranches, nil
}
