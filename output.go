package main

import (
	"log"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/olekukonko/tablewriter"
)

func printSummary(stalePrs map[string][]*github.PullRequest, staleBranches map[string][]*Branch,
	alertPrs map[string][]*github.PullRequest, alertBranches map[string][]*Branch) {

	//Stale PRs
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Number", "Title", "Last Updated", "Link"})

	for user, prs := range stalePrs {
		for _, pr := range prs {
			table.Append([]string{user, strconv.Itoa(*pr.Number), *pr.Title, pr.UpdatedAt.String(), *pr.URL})
		}
	}
	log.Printf("1. Stale Pull requests (closure pending)")
	table.Render()

	//Alert PRs
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Number", "Title", "Last Updated", "Link"})

	for user, prs := range alertPrs {
		for _, pr := range prs {
			table.Append([]string{user, strconv.Itoa(*pr.Number), *pr.Title, pr.UpdatedAt.String(), *pr.URL})
		}
	}
	log.Printf("2. Pull requests reaching stale")
	table.Render()

	//Stale Branches
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Branch", "Last Updated", "Link"})

	for user, branches := range staleBranches {
		for _, branch := range branches {
			table.Append([]string{user, *branch.Name, branch.ExCommit.Committer.Date.String(), *branch.Commit.HTMLURL})
		}
	}
	log.Printf("3. Stale Branches")
	table.Render()

	//Alert Branches
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Branch", "Last Updated", "Link"})

	for user, branches := range alertBranches {
		for _, branch := range branches {
			table.Append([]string{user, *branch.Name, branch.ExCommit.Committer.Date.String(), *branch.Commit.HTMLURL})
		}
	}
	log.Printf("4. Branches reaching stale")
	table.Render()
}
