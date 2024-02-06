package jiraissue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
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
	Parent      *FieldProps   `json:"parent,omitempty"` // EpicID
	Labels      []string      `json:"labels"`

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

// CreateJiraIssue creates an issue in Jira
func CreateJiraIssue(summary, timeEstimate, description, epicKey, issueType, priorityID, assigneeID string, components, labels []string, isDebugEnabled bool) (string, error) {
	// [ ] TODO: refactor params into an options struct for easier arguments management and validate this env vars at cmd.
	jiraProjectKey, ok := os.LookupEnv("JIRA_PROJECT_KEY")
	if !ok {
		log.Fatal("Environment Variable JIRA_PROJECT_KEY not set")
	}
	jiraAPIToken, ok := os.LookupEnv("JIRA_API_TOKEN")
	if !ok {
		log.Fatal("Environment Variable JIRA_API_TOKEN not set")
	}

	issue := JiraIssue{
		Fields: Fields{
			Summary: summary,
			Description: &ContentDef{
				Content: []*ContentDef{
					{
						Content: []*ContentDef{
							{
								Text: description,
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
				Name: issueType,
			},
			Project: Project{
				Key: jiraProjectKey,
			},
			Priority: FieldProps{
				ID: priorityID,
			},
			Assignee: FieldProps{
				ID: assigneeID,
			},
			Parent: &FieldProps{
				Key: epicKey,
			},
			// TimeTracking: &TimeTracking{
			// 	OriginalEstimate: timeEstimate,
			// },
		},
	}

	issueComponents := []*FieldProps{}
	for _, c := range components {
		issueComponents = append(issueComponents, &FieldProps{Name: c})
	}

	issueLabels := []string{}
	for _, l := range labels {
		issueLabels = append(issueLabels, l)
	}

	issue.Fields.Components = issueComponents
	issue.Fields.Labels = issueLabels

	issueJSON, err := json.Marshal(issue)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://pagerduty.atlassian.net/rest/api/3/issue", bytes.NewBuffer(issueJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+jiraAPIToken)

	if isDebugEnabled {
		dumpedReq, err := httputil.DumpRequest(req, true)
		if err == nil {
			log.Println("[DEBUG] Request::>", string(dumpedReq))
			log.Println("---")
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if isDebugEnabled {
		dumpedResp, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[DEBUG] Response::>", string(dumpedResp))
			log.Println("---")
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
