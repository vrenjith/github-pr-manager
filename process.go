package main

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/github"
)

// PR Handling
// 1. Every PR which is older than 7 days, summarized in a single Email
// 2. Every PR which is not updated for more than 14 days will be closed automatically and branch deleted.

// Branch handling
// 1. Every branch that:
// - does not have a PR and
// - without branch protection and
// - is at least 7 days old
// will be summarized in the email
// 2. Every branch that does not have a PR and is at least 14 days old will be deleted after creating a dummy PR
// - does not have a PR and
// - without branch protection and
// - is at least 14 days old
// will be deleted after creating a dummy PR

func analysePrs(pulls []*github.PullRequest, stale map[string][]*github.PullRequest, alert map[string][]*github.PullRequest, args *Arguments) {
	for _, pull := range pulls {
		log.Println("Pull:", *pull.ID, *pull.Number, *pull.Title)

		durationSinceLastUpdate := int(time.Since(*pull.UpdatedAt).Hours())
		user := *pull.User.Login

		if durationSinceLastUpdate > args.prStaleDays*24 {
			userprs, ok := stale[user]
			if !ok {
				userprs = make([]*github.PullRequest, 0)
			}
			stale[user] = append(userprs, pull)
		} else if durationSinceLastUpdate > (args.prStaleDays*24 - args.alertDays*24) {
			userprs, ok := alert[user]
			if !ok {
				userprs = make([]*github.PullRequest, 0)
			}
			alert[user] = append(userprs, pull)
		}
	}
}

// Branch is our wrapper for the actual branch
// that includes the commit details too
type Branch struct {
	*github.Branch
	ExCommit *github.Commit
}

func analyseBranches(client *github.Client, repo *github.Repository, branches []*github.Branch,
	stale map[string][]*Branch, alert map[string][]*Branch, args *Arguments) {
	for _, branch := range branches {
		log.Println("Branch:", *branch.Name)

		if _, ok := args.ignoreBranchesMap[*branch.Name]; ok {
			log.Println("Ignoring branch:", *branch.Name)
			continue
		}

		if *branch.Protected {
			log.Println("Ignoring protected branch:", *branch.Name)
			continue
		}

		commit, _, _ := client.Git.GetCommit(context.Background(), *repo.Owner.Login, *repo.Name, *branch.Commit.SHA)

		exBranch := Branch{branch, commit}

		durationSinceLastUpdate := int(time.Since(*commit.Author.Date).Hours())

		user := *commit.Author.Email

		if durationSinceLastUpdate > args.branchStaleDays*24 {
			userbranches, ok := stale[user]
			if !ok {
				userbranches = make([]*Branch, 0)
			}
			stale[user] = append(userbranches, &exBranch)
		} else if durationSinceLastUpdate > (args.branchStaleDays*24 - args.alertDays*24) {
			userbranches, ok := alert[user]
			if !ok {
				userbranches = make([]*Branch, 0)
			}
			alert[user] = append(userbranches, &exBranch)
		}
	}
}
