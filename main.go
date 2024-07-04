package jiraissue

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
)

// JiraIssue represents the structure of the issue to be created
type JiraIssue struct {
	Fields Fields `json:"fields"`
}

// Fields represent the Jira issue fields
type Fields struct {
	Summary     string        `json:"summary"`
	Description *ContentDef   `json:"description,omitempty"`
	Components  []*FieldProps `json:"components,omitempty"`
	Issuetype   Issuetype     `json:"issuetype"`
	Project     Project       `json:"project"`
	Priority    FieldProps    `json:"priority"`
	Assignee    FieldProps    `json:"assignee"`
	Parent      *FieldProps   `json:"parent,omitempty"` // EpicKey
	Labels      []string      `json:"labels"`
	FixVersions []*FixVersion `json:"fixVersions,omitempty"`

	// To use this setting is needed to follow the steps here https://community.atlassian.com/t5/Jira-Content-Archive-questions/Unable-to-set-TimeTracking-using-JIRA-API/qaq-p/1917217#M246455
	TimeTracking *TimeTracking `json:"timetracking,omitempty"`
}

type FieldProps struct {
	Key  string `json:"key,omitempty"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Issuetype struct {
	Name string `json:"name"`
}

type Project struct {
	Key string `json:"key"`
}

type TimeTracking struct {
	OriginalEstimate  string `json:"originalEstimate,omitempty"`
	RemainingEstimate string `json:"remainingEstimate,omitempty"`
}

type ContentDef struct {
	Type    string        `json:"type"`
	Content []*ContentDef `json:"content,omitempty"`
	Text    string        `json:"text,omitempty"`
	Version int           `json:"version,omitempty"`
}

type Client struct {
	httpClient *http.Client
	APIToken   string
	Debug      bool
	DryRun     bool
}

type FixVersion struct {
	Self        string `json:"self,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Id          string `json:"id,omitempty"`
	Archived    *bool  `json:"archived,omitempty"`
	Released    *bool  `json:"released,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// CreateJiraIssue creates an issue in Jira
func CreateJiraIssue(summary, timeEstimate, description, epicKey, issueType, priorityID, assigneeID, fixVersionName, csvPath string, components, labels []string, isDebugEnabled, isDryRunning bool) error {
	// [ ] TODO: refactor params into an options struct for easier arguments management and validate this env vars at cmd.
	jiraProjectKey, ok := os.LookupEnv("JIRA_PROJECT_KEY")
	if !ok {
		log.Fatal("Environment Variable JIRA_PROJECT_KEY not set")
	}
	jiraAPIToken, ok := os.LookupEnv("JIRA_API_TOKEN")
	if !ok {
		log.Fatal("Environment Variable JIRA_API_TOKEN not set")
	}

	var issues = make([]*JiraIssue, 0)

	if csvPath != "" {
		var errCSV error
		issues, errCSV = expandIssuesFromCSV(csvPath, priorityID, issueType, jiraProjectKey, assigneeID)
		if errCSV != nil {
			return errCSV
		}
	}

	if csvPath == "" && summary == "" {
		log.Fatal("When creating single Jira issues --summary flag is required")
	}

	if csvPath == "" && summary != "" {
		issues = append(issues, createIssuePayload(IssuePayloadContent{
			summary:        summary,
			description:    description,
			issueType:      issueType,
			jiraProjectKey: jiraProjectKey,
			priorityID:     priorityID,
			assigneeID:     assigneeID,
			epicKey:        epicKey,
			timeEstimate:   timeEstimate,
			components:     components,
			labels:         labels,
			fixVersionName: fixVersionName,
		}))
	}

	c := &Client{
		httpClient: &http.Client{},
		APIToken:   jiraAPIToken,
		Debug:      isDebugEnabled,
		DryRun:     isDryRunning,
	}

	ctx := context.Background()
	err := handleJiraIssuesCreation(ctx, c, issues)
	if err != nil {
		return err
	}

	return nil
}

func createJiraIssueAPICall(ctx context.Context, c *Client, issue *JiraIssue) (string, error) {
	issueJSON, err := json.Marshal(issue)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://pagerduty.atlassian.net/rest/api/3/issue", bytes.NewBuffer(issueJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.APIToken)

	// [ ] TODO: Implement logging in roundtripper
	if c.Debug || c.DryRun {
		dumpedReq, err := httputil.DumpRequest(req, true)
		if err == nil {
			log.Println("[DEBUG] Request::>", string(dumpedReq))
		}
	}

	if c.DryRun {
		return fmt.Sprintf("%s-dry-run", issue.Fields.Project.Key), nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// [ ] TODO: Implement logging in roundtripper
	if c.Debug {
		dumpedResp, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[DEBUG] Response::>", string(dumpedResp))
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		if errors, ok := result["errors"]; ok {
			errMessage, _ := json.MarshalIndent(errors, "", "  ")
			return "", fmt.Errorf("Errors: %v", string(errMessage))
		}
	}

	issueID, ok := result["key"].(string)
	if !ok {
		return "", fmt.Errorf("error parsing issue ID from response")
	}

	return issueID, nil
}

func handleJiraIssuesCreation(ctx context.Context, c *Client, issues []*JiraIssue) (err error) {
	var wg sync.WaitGroup

	jirasubdomain := os.Getenv("JIRA_SUBDOMAIN")

	for _, iss := range issues {
		wg.Add(1)

		go func(issue *JiraIssue) {
			defer wg.Done()

			var issueKey string
			issueKey, err = createJiraIssueAPICall(ctx, c, issue)

			var successMsg string
			if jirasubdomain == "" {
				successMsg = fmt.Sprintf("Issue created. %s", issueKey)
			}
			successMsg = fmt.Sprintf("Issue created. Link to issue https://%s.atlassian.net/browse/%s", jirasubdomain, issueKey)
			fmt.Println(successMsg)
		}(iss)

		if err != nil {
			return err
		}
	}

	wg.Wait()

	return err
}

type IssuePayloadContent struct {
	summary        string
	description    string
	issueType      string
	jiraProjectKey string
	priorityID     string
	assigneeID     string
	epicKey        string
	timeEstimate   string
	components     []string
	labels         []string
	fixVersionName string
}

func createIssuePayload(c IssuePayloadContent) *JiraIssue {
	var p JiraIssue = JiraIssue{
		Fields: Fields{
			Summary: c.summary,
			Description: &ContentDef{
				Content: []*ContentDef{
					{
						Content: []*ContentDef{
							{
								Text: c.description,
								Type: "text",
							},
						},
						Type: "paragraph",
					},
				},
				Type:    "doc",
				Version: 1,
			},
			Issuetype: Issuetype{
				Name: c.issueType,
			},
			Project: Project{
				Key: c.jiraProjectKey,
			},
			Priority: FieldProps{
				ID: c.priorityID,
			},
			Assignee: FieldProps{
				ID: c.assigneeID,
			},
			Parent: &FieldProps{
				Key: c.epicKey,
			},
			// TimeTracking: &TimeTracking{
			// 	OriginalEstimate: c.timeEstimate,
			// },
			FixVersions: []*FixVersion{
				{
					Name: c.fixVersionName,
				},
			},
		},
	}

	issueComponents := []*FieldProps{}
	for _, c := range c.components {
		issueComponents = append(issueComponents, &FieldProps{Name: c})
	}

	issueLabels := []string{}
	for _, l := range c.labels {
		issueLabels = append(issueLabels, l)
	}

	p.Fields.Components = issueComponents
	p.Fields.Labels = issueLabels

	return &p
}

func expandIssuesFromCSV(csvPath, priorityID, issueType, jiraProjectKey, assigneeID string) ([]*JiraIssue, error) {
	var issues = make([]*JiraIssue, 0)

	csvBytes, err := os.ReadFile(csvPath)
	if err != nil {
		return nil, err
	}

	csvContent := string(csvBytes)
	r := csv.NewReader(strings.NewReader(csvContent))
	r.Comma = ';'
	r.Comment = '#'
	r.TrimLeadingSpace = true

	for i := 0; i >= 0; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if i < 1 {
			continue
		}

		// summary;description;time;epic;components;labels;fixVersionName
		var (
			summary        = record[0]
			description    = record[1]
			timeEstimate   = record[2]
			epic           = record[3]
			components     = expandListedCellValues(record[4])
			labels         = expandListedCellValues(record[5])
			fixVersionName = record[6]
		)

		issues = append(
			issues,
			createIssuePayload(IssuePayloadContent{
				summary:        summary,
				description:    description,
				issueType:      issueType,
				jiraProjectKey: jiraProjectKey,
				priorityID:     priorityID,
				assigneeID:     assigneeID,
				epicKey:        epic,
				timeEstimate:   timeEstimate,
				components:     components,
				labels:         labels,
				fixVersionName: fixVersionName,
			}),
		)
	}

	return issues, nil
}

func expandListedCellValues(r string) []string {
	v := strings.Split(r, ",")
	var result []string

	for _, c := range v {
		result = append(result, strings.TrimSpace(c))
	}

	return result
}
