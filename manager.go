package main

import (
	"context"
	"flag"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
)

/*
// PR Handling
1. Every PR which is older than 7 days, summarized in a single Email
2. Every PR which is not updated for more than 14 days will be closed automatically and branch deleted.

// Branch handling
1. Every branch that:
- does not have a PR and
- without branch protection and
- is at least 7 days old
will be summarized in the email
2. Every branch that does not have a PR and is at least 14 days old will be deleted after creating a dummy PR
- does not have a PR and
- without branch protection and
- is at least 14 days old
will be deleted after creating a dummy PR

Mail format
1. Mails will be send as a summary per user across PRs and Branches
2. Two sections, one for branches and one for PRs
3. Each row is hyper linked
4. Each row says the number of days before deletion

*/

type Arguments struct {
    owners  string
    token string
    apiUrl string
    ignoreBranches string
    repoPattern string
    prStaleDays int
    branchStaleDays int
    closePrs bool
    deleteBranches bool
    detectJira bool
    jiraUserName string
    jiraPassword string
    sendEmails bool
    smtpServer string
    emailDomain string
}

func main() {

	ag := handleArguments()

	client, _ := getGithubClient()
	//owners := "Ariba,Ariba-cobalt"
	orgs := strings.Split(ag.owners, ",")
	for _, org := range orgs {
		log.Println("Checking repositories under org:", org)
		repos, _ := getOrgRepos(client, org)

		for _, repo := range repos {
			log.Println("Checking pull-requests under repo:", *repo.Name)
			pulls, _ := getRepoPulls(client, org, *repo.Name)
			for _, pull := range pulls {
				log.Println("Pull:", *pull.ID, *pull.Number, *pull.Title)
			}

			log.Println("Checking branches under repo:", *repo.Name)
			branches, _ := getRepoBranches(client, org, *repo.Name)
			for _, branch := range branches {
				log.Println("Branch:", *branch.Name)
			}

		}
	}

	log.Println("Main complete")
}

func handleArguments() Arguments {
    ag := Arguments{}

    flag.StringVar(&ag.owners, "owners", "", "Organization(s) (comma seperated) to check (required)")
    flag.StringVar(&ag.token, "token", os.Getenv("GITHUB_TOKEN"), "Authentication token (required)(optional if GITHUB_TOKEN environment variable is set)")
    flag.StringVar(&ag.apiUrl, "api-url", os.Getenv("GITHUB_API_URL"), "Github API URL. (required)(optional if GITHUB_API_URL environment variable is set)")
    flag.StringVar(&ag.ignoreBranches, "ignore-branches", "master,develop", "Branches to ignore (comma seperated) (optional)")
    flag.StringVar(&ag.repoPattern, "repo-pattern", ".*", "Repository pattern to filter repositories (optional)")

    flag.IntVar(&ag.prStaleDays, "pr-stale-days", 7, "Number of inactive days to consider a PR as stale (optional)")
    flag.IntVar(&ag.branchStaleDays, "branch-stale-days", 14, "Number of inactive days to consider a PR as stale (optional)")

    flag.BoolVar(&ag.closePrs, "close-prs", true, "Close the stale pull requests which has crossed the pr-stale-days value")
    flag.BoolVar(&ag.deleteBranches, "delete-branches", false, "Delete the stale branches which has crossed the branch-stale-days value")
    flag.BoolVar(&ag.detectJira, "detect-jira", false, "Attempt to detect JIRA ID for the pull requests and branches from their names and automatically add a comment (See also jira-user-name & jira-password)")
    flag.BoolVar(&ag.sendEmails, "send-emails", false, "Send summary emails to each committer about stale pull requests and branches")
    flag.StringVar(&ag.smtpServer, "smtp-server", "", "SMTP Server to use (authentication not supported) (optional) (Mandatory if send-emails is set to true)")
    flag.StringVar(&ag.emailDomain, "email-domain", "", "Email domain to be used (optional)(Mandatory if send-emails is set to true)")

    flag.StringVar(&ag.jiraUserName, "jira-user-name", os.Getenv("JIRA_USERNAME"), "JIRA user name for commenting (optional)(Mandatory if detect-jira is set to true)(respects JIRA_USERNAME environment variable)")
    flag.StringVar(&ag.jiraPassword, "jira-password", os.Getenv("JIRA_PASSWORD"), "JIRA password for commenting (optional)(Mandatory if detect-jira is set to true)(respects JIRA_PASSWORD environment variable)")

    flag.Parse()

    if len(ag.token) == 0 {
        logAndExit("Github API Token")
    }
    if len(ag.apiUrl) == 0 {
        logAndExit("Github API URL")
    }
    if ag.detectJira && (len(ag.jiraUserName) == 0 || len(ag.jiraPassword)  == 0) {
        logAndExit("JIRA User name and password")
    }
    if ag.sendEmails && (len(ag.smtpServer) == 0 || len(ag.emailDomain)  == 0) {
        logAndExit("SMTP Server and Email Domain")
    }
    return ag
}

func logAndExit(message string) {
    log.Println(message, "needs to be specified. See help below.")
    flag.Usage()
    os.Exit(1)
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

func getRepoBranches(client *github.Client, owner string, repo string) ([]*github.Branch, error) {
	var allBranches []*github.Branch
	opt := &github.BranchListOptions{}
	for {
		branches, resp, err := client.Repositories.ListBranches(context.Background(), owner, repo, opt)
		if err != nil {
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
