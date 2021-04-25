package main

import (
	"fmt"
	"github.com/StevenACoffman/workon/pkg/atlassian"
	"github.com/andygrunwald/go-jira"
	"os"
	"strings"

	"github.com/StevenACoffman/workon/pkg/git"
)

const (
	// exitFail is the exit code if the program fails.
	exitFail = 1
	// exitSuccess is the exit code if the program succeeds
	exitSuccess = 0
)

func main() {
	config := atlassian.ConfigureJira()
	args := getArgs()
	if len(args) == 0 {
		fmt.Println("You need to specify a Jira issue or branch name")
		os.Exit(exitFail)
	}
	prefix := getEnv("GIT_WORKON_PREFIX", "feature/")
	branchName := ParseTopicBranch(args[0], config.Host)
	fullBranchName := fmt.Sprintf("%s%s", prefix, branchName)

	err := git.CreateBranch(fullBranchName)
	if err != nil {
		fmt.Printf("failed to create branch %s: %s\n", fullBranchName, err)
		os.Exit(exitFail)
	}

	tp := jira.BasicAuthTransport{
		Username: config.User,
		Password: config.Token,
	}
	issueKey := atlassian.ParseJiraIssue(args[0], config.Host)

	jiraClient, err := jira.NewClient(tp.Client(), config.Host)
	if err != nil {
		panic(err)
	}

	issue, _, issueErr := jiraClient.Issue.Get(issueKey, nil)
	if issueErr != nil {
		fmt.Printf("Unable to get Issue %s: %+v", issueErr)
		os.Exit(exitFail)
	}
	err = atlassian.AssignIssueToSelf(jiraClient, issue, issueKey)
	if err != nil {
		fmt.Println(err)
		os.Exit(exitFail)
	}
	currentStatus := issue.Fields.Status.Name

	err = atlassian.MoveIssueToInProgress(jiraClient, issue, currentStatus, issueKey)
	if err != nil {
		fmt.Println(err)
		os.Exit(exitFail)
	}
	os.Exit(exitSuccess)
}

// getArgs - skips binary name, and no flags please
func getArgs() []string {
	argSkip := 1 // skip binary name if there was one
	if len(os.Args) <= argSkip {
		return []string{}
	}

	var args []string
	for _, arg := range os.Args[argSkip:] {
		// skip flags
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			continue
		}
		args = append(args, arg)
	}
	return args
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// ParseTopicBranch - Sanitizes input
//  + Trims leading and trailing whitespace
//  + Trims a JIRA browse URL
// Note: Leaves anything after a JIRA issue
// e.g. "TEAM-1234/part1" is fine
func ParseTopicBranch(issueKey , host string) string {
	issueKey = strings.TrimSpace(issueKey)
	if issueKey == "" {
		return issueKey
	}
	if strings.HasPrefix(issueKey, host) {
		issueKey = strings.TrimPrefix(issueKey, host)
		issueKey = strings.TrimPrefix(issueKey, "/browse/")
	}
	return issueKey
}
