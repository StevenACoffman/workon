package atlassian

import (
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

const InProgressStatusID = "10105"

func MoveIssueToInProgress(jiraClient *jira.Client, issue *jira.Issue, originalStatus string, issueKey string) error {
	if issue.Fields.Status.ID != InProgressStatusID {
		// Transition issue to in progress / start development / resume development
		err := startIssue(jiraClient, issueKey)
		if err != nil {
			return err
		}
		issue, _, err = jiraClient.Issue.Get(issueKey, nil)
		if err != nil {
			return err
		}
		fmt.Printf("Issue %s Status changed status from: %s set to: %+v\n",
			issueKey, originalStatus, issue.Fields.Status.Name)
	} else {
		fmt.Println("Already In Progress")
	}
	return nil
}

func AssignIssueToSelf(jiraClient *jira.Client, issue *jira.Issue, issueKey string) error {
	self, _, selfErr := jiraClient.User.GetSelf()
	if selfErr != nil {
		return fmt.Errorf("unable to get myself %s: %+v", selfErr)
	}

	if issue.Fields.Assignee == nil || self.AccountID != issue.Fields.Assignee.AccountID {
		_, assignErr := jiraClient.Issue.UpdateAssignee(issueKey, self)
		if assignErr != nil {
			return fmt.Errorf("unable to assign to yourself %s: %+v", assignErr)
		}
		fmt.Printf("Re-Assigned %s from %s\n", issueKey, DisplayJiraUser(issue.Fields.Assignee))
	} else {
		fmt.Println("Already assigned to to you")
	}
	return nil
}

// ParseJiraIssue - Sanitizes input
//  + Trims leading and trailing whitespace
//  + Trims a browse URL
//  + Trims anything after ABCD-1234
func ParseJiraIssue(issueKey, host string) string {
	issueKey = strings.TrimSpace(issueKey)
	if issueKey == "" {
		return issueKey
	}
	if strings.HasPrefix(issueKey, host) {
		issueKey = strings.TrimPrefix(issueKey, host)
		issueKey = strings.TrimPrefix(issueKey, "/browse/")
	}
	// This will remove everything after the ABCD-1234
	reg := regexp.MustCompile(`([A-Za-z]+-[0-9]+).*`)
	res := reg.ReplaceAllString(issueKey, "${1}")
	return res
}

func DisplayJiraUser(jiraUser *jira.User) string {
	if jiraUser == nil {
		return "Unassigned"
	}
	return jiraUser.DisplayName + " (" + jiraUser.EmailAddress + ")"
}

func startIssue(jiraClient *jira.Client, issueKey string) error {
	var transitionID string
	possibleTransitions, _, err := jiraClient.Issue.GetTransitions(issueKey)
	if err != nil {
		return err
	}

	for _, v := range possibleTransitions {
		if isInProgressTransition(v) {
			transitionID = v.ID
			break
		}
	}

	_, err = jiraClient.Issue.DoTransition(issueKey, transitionID)
	return err
}

func isInProgressTransition(v jira.Transition) bool {
	inProgressTransitions := []string{"In Progress","Resume Development","Start Development"}
	for _, transName := range inProgressTransitions {
		if strings.EqualFold(v.Name, transName) {
			return true
		}
	}
	return false
}

// Config struct
type Config struct {
	Host  string `json:"host"`
	User  string `json:"user"`
	Token string `json:"token"`
}

// ReadConfigFromFile returns an error if file does not exist
func ReadConfigFromFile() (*Config, error) {
	configFile, configErr := expandTilde(getEnv("ATLASSIAN_CONFIG_FILE", "~/.config/jira"))

	if configErr != nil {
		// if we can't get the config file, then we have no hope.
		return nil, fmt.Errorf("unable to get config file directory +v", configErr)
	}

	var config Config
	configJSON, err := ioutil.ReadFile(configFile)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(configJSON, &config)

	if err != nil {
		return &config, err
	}

	config.Token = getEnv("ATLASSIAN_API_TOKEN", config.Token)
	config.Host = getEnv("ATLASSIAN_HOST", config.Host)
	config.User = getEnv("ATLASSIAN_API_USER", config.User)

	return &config, nil
}

func ReadConfigFromEnv() *Config {
	host := os.Getenv("ATLASSIAN_HOST")
	user := os.Getenv("ATLASSIAN_API_USER")
	token := os.Getenv("ATLASSIAN_API_TOKEN")
	config := Config{
		Host:  host,
		User:  user,
		Token: token,
	}
	return &config
}

func ConfigureJira() *Config {
	config, err := ReadConfigFromFile()
	if err != nil {
		// we got an error reading from the config file, so just use env
		return ReadConfigFromEnv()
	}
	// allow env to replace file config
	config.Token = getEnv("ATLASSIAN_API_TOKEN", config.Token)
	config.Host = getEnv("ATLASSIAN_HOST", config.Host)
	config.User = getEnv("ATLASSIAN_API_USER", config.User)
	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// "~/.gitignore" -> "/home/tyru/.gitignore"
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	var paths []string
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	for _, p := range strings.Split(path, string(filepath.Separator)) {
		if p == "~" {
			paths = append(paths, u.HomeDir)
		} else {
			paths = append(paths, p)
		}
	}
	return "/" + filepath.Join(paths...), nil
}
