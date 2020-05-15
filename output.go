package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"text/template"

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
			table.Append([]string{user, *branch.Name, branch.ExCommit.Committer.Date.String(), *branch.Commit.URL})
		}
	}
	log.Printf("3. Stale Branches")
	table.Render()

	//Alert Branches
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Branch", "Last Updated", "Link"})

	for user, branches := range alertBranches {
		for _, branch := range branches {
			table.Append([]string{user, *branch.Name, branch.ExCommit.Committer.Date.String(), *branch.Commit.URL})
		}
	}
	log.Printf("4. Branches reaching stale")
	table.Render()
}

func emailSummary(ag Arguments, stalePrs map[string][]*github.PullRequest, staleBranches map[string][]*Branch,
	alertPrs map[string][]*github.PullRequest, alertBranches map[string][]*Branch) {

	allUsers := make([]string, 0)
	for k := range stalePrs {
		allUsers = append(allUsers, k)
	}
	for k := range alertPrs {
		allUsers = append(allUsers, k)
	}
	for _, user := range allUsers {
		sprs, ok := stalePrs[user]
		if !ok {
			sprs = make([]*github.PullRequest, 0)
		}
		aprs, ok := alertPrs[user]
		if !ok {
			aprs = make([]*github.PullRequest, 0)
		}
		if len(sprs) > 0 || len(aprs) > 0 {
			//sendMail(ag, user, sprs, aprs)
		}
	}
	//tmpl, _ := template.ParseFiles("email-template.html")
	//tmpl.Execute()
}

// EmailData holds the info to be rendered for the HTML email to the user
type EmailData struct {
	stalePrs      []*github.PullRequest
	alertPrs      []*github.PullRequest
	staleBranches []*Branch
	alertBranches []*Branch
}

func sendMail(args *Arguments, user string, category string, data *EmailData) {
	tmpl := template.New("email-template")
	tmpl, _ = tmpl.ParseFiles("email-template.html")

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		log.Println("Mail template rendering failed")
		return
	}
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subject := "Subject: Notification - " + category + "\n"
	msg := []byte(subject + mime + "\n" + buf.String())

	smtp.SendMail(args.smtpServer, smtp.PlainAuth("", "", "", args.smtpServer),
		args.fromEmail, []string{fmt.Sprintf("%s@%s", user, args.emailDomain)}, msg)
}
