package main

import (
	"log"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/olekukonko/tablewriter"
)

func printSummary(stalePrs map[string][]*github.PullRequest, staleBranches map[string][]*github.Branch,
	alertPrs map[string][]*github.PullRequest, alertBranches map[string][]*github.Branch) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Number", "Title", "Last Updated"})

	for user, prs := range stalePrs {
		for _, pr := range prs {
			table.Append([]string{user, strconv.Itoa(*pr.Number), *pr.Title, pr.UpdatedAt.String()})
		}
	}
	log.Printf("1. Stale Pull requests")
	table.Render()
}
