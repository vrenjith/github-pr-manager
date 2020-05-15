package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

// Mail format
// 1. Mails will be send as a summary per user across PRs and Branches
// 2. Two sections, one for branches and one for PRs
// 3. Each row is hyper linked
// 4. Each row says the number of days before deletion

// Arguments is used to collect all the
// command line arguments being passed to the utility
type Arguments struct {
	owners            string
	token             string
	apiURL            string
	ignoreBranches    string
	ignoreBranchesMap map[string]struct{}
	repoPattern       string
	prStaleDays       int
	branchStaleDays   int
	closePrs          bool
	deleteBranches    bool
	detectJira        bool
	jiraUserName      string
	jiraPassword      string
	sendEmails        bool
	smtpServer        string
	emailDomain       string
	alertDays         int
	adminEmail        string
	fromEmail         string
}

func main() {

	ag := handleArguments()

	stalePrs := make(map[string][]*github.PullRequest)
	staleBranches := make(map[string][]*Branch)
	alertPrs := make(map[string][]*github.PullRequest)
	alertBranches := make(map[string][]*Branch)

	client, _ := getGithubClient(ag)
	orgs := strings.Split(ag.owners, ",")
	for _, org := range orgs {
		log.Println("Checking repositories under org:", org)
		repos, _ := getOrgRepos(client, org)

		for _, repo := range repos {
			log.Println("Checking pull-requests under repo:", *repo.Name)
			pulls, _ := getRepoPulls(client, org, *repo.Name)
			analysePrs(pulls, stalePrs, alertPrs, &ag)

			log.Println("Checking branches under repo:", *repo.Name)
			branches, _ := getRepoBranches(client, org, *repo.Name)
			analyseBranches(client, repo, branches, staleBranches, alertBranches, &ag)
		}
	}
	printSummary(stalePrs, staleBranches, alertPrs, alertBranches)

	emailSummary(ag, stalePrs, staleBranches, alertPrs, alertBranches)

	log.Println("Main complete")
}

func handleArguments() Arguments {
	ag := Arguments{}

	flag.StringVar(&ag.owners, "owners", "", "Organization(s) (comma seperated) to check (required)")
	flag.StringVar(&ag.token, "token", os.Getenv("GITHUB_TOKEN"), "Authentication token (required)(optional if GITHUB_TOKEN environment variable is set)")
	flag.StringVar(&ag.apiURL, "api-url", os.Getenv("GITHUB_API_URL"), "Github API URL. (required)(optional if GITHUB_API_URL environment variable is set)")
	flag.StringVar(&ag.ignoreBranches, "ignore-branches", "master,develop", "Branches to ignore (comma seperated) (optional)")
	flag.StringVar(&ag.repoPattern, "repo-pattern", ".*", "Repository pattern to filter repositories (optional)")

	flag.IntVar(&ag.prStaleDays, "pr-stale-days", 14, "Number of inactive days to consider a PR as stale (optional)")
	flag.IntVar(&ag.branchStaleDays, "branch-stale-days", 14, "Number of inactive days to consider a PR as stale (optional)")
	flag.IntVar(&ag.alertDays, "alert-days", 14, "Number of days in advance to start alerting about stale branches and pull requests (optional)")

	flag.BoolVar(&ag.closePrs, "close-prs", true, "Close the stale pull requests which has crossed the pr-stale-days value")
	flag.BoolVar(&ag.deleteBranches, "delete-branches", false, "Delete the stale branches which has crossed the branch-stale-days value")
	flag.BoolVar(&ag.detectJira, "detect-jira", false, "Attempt to detect JIRA ID for the pull requests and branches from their names and automatically add a comment (See also jira-user-name & jira-password)")
	flag.BoolVar(&ag.sendEmails, "send-emails", false, "Send summary emails to each committer about stale pull requests and branches")
	flag.StringVar(&ag.smtpServer, "smtp-server", "", "SMTP Server to use (authentication not supported) (optional) (Mandatory if send-emails is set to true)")
	flag.StringVar(&ag.emailDomain, "email-domain", "", "Email domain to be used (optional)(Mandatory if send-emails is set to true)")
	flag.StringVar(&ag.adminEmail, "admin-email", "", "Email address of the admin who will receive all emails. (optional)(Mandatory if send-emails is set to true)")
	flag.StringVar(&ag.fromEmail, "from-email", "", "From email address for all emails. (optional)(Mandatory if send-emails is set to true)")

	flag.StringVar(&ag.jiraUserName, "jira-user-name", os.Getenv("JIRA_USERNAME"), "JIRA user name for commenting (optional)(Mandatory if detect-jira is set to true)(respects JIRA_USERNAME environment variable)")
	flag.StringVar(&ag.jiraPassword, "jira-password", os.Getenv("JIRA_PASSWORD"), "JIRA password for commenting (optional)(Mandatory if detect-jira is set to true)(respects JIRA_PASSWORD environment variable)")

	flag.Parse()

	if len(ag.token) == 0 {
		logAndExit("Github API Token")
	}
	if len(ag.apiURL) == 0 {
		logAndExit("Github API URL")
	}
	if ag.detectJira && (len(ag.jiraUserName) == 0 || len(ag.jiraPassword) == 0) {
		logAndExit("JIRA User name and password")
	}
	if ag.sendEmails && len(ag.smtpServer) == 0 {
		logAndExit("SMTP Server")
	}
	if ag.sendEmails && len(ag.emailDomain) == 0 {
		logAndExit("Email Domain")
	}
	if ag.sendEmails && len(ag.adminEmail) == 0 {
		logAndExit("Administrator Email")
	}
	if ag.sendEmails && len(ag.fromEmail) == 0 {
		logAndExit("From Email")
	}
	if len(ag.owners) == 0 {
		logAndExit("Github organizations/owners")
	}
	//https://www.davidkaya.com/sets-in-golang/
	ag.ignoreBranchesMap = make(map[string]struct{})
	for _, branch := range strings.Split(ag.ignoreBranches, ",") {
		ag.ignoreBranchesMap[branch] = struct{}{}
	}
	return ag
}

func logAndExit(message string) {
	log.Println(message, "needs to be specified. See help below.")
	flag.Usage()
	os.Exit(1)
}
