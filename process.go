package main

import (
	"log"

	"github.com/google/go-github/github"
)

func buildStalePrList(pulls []*github.PullRequest, result map[string][]*github.PullRequest) {
	for _, pull := range pulls {
		log.Println("Pull:", *pull.ID, *pull.Number, *pull.Title)
	}
}

func buildStaleBranchList(branches []*github.Branch, result map[string][]*github.Branch) {
	for _, branch := range branches {
		log.Println("Branch:", *branch.Name)
	}

}
