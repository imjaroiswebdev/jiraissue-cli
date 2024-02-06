# Jira issue creator cli

## Description

This tool is a CLI application for creating issues in Jira based on a specific workflow a follow to manage Jira issues, at this point it is not intended to be of general use for the majority of cases.

## Installation
1. Clone the repository.
2. Build the application: `go build`

## Configuration
Set the following environment variables:
- `JIRA_API_TOKEN`: Your Jira API token. [(How to get one)](https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/)
- `JIRA_PROJECT_KEY`: The key of your Jira project. (e.g., AWESOME)
- `JIRA_SUBDOMAIN`: (Optional) Your Jira subdomain. (e.g., `https://<subdomain>.atlassian.net`)
- `JIRA_ASSIGNEE_ID`: Jira issue assignee id. (e.g., `62703a40ca1fae106ae98fed`)

## Usage
Run the application with the necessary parameters:

```bash
$ go run ./cmd/jiraissue.go --help
Jira CLI to create issues

Usage:
  jiraissue [flags]

Flags:
      --assignee string         Issue assignee id
  -c, --component stringArray   Components names separated list
      --debug                   Enable debug of API calls
  -d, --description string      Description of the issue
  -e, --epic string             Epic ID
  -h, --help                    help for jiraissue
  -l, --label stringArray       Labels names separated list
  -s, --summary string          Summary of the issue (required)
  -t, --time string             Time estimation

go run ./cmd \
  --summary "New Awesome created through jiraissue cli" \
  --time "4h" \
  --description 'Issue created while testing `jiraissue` cli App' \
  --epic "AWESOME-1758" \
  --component "BACKEND" \
  --component "MIDDLEWARE" \
  --component "FRONTEND" \
  --label "AWESOME_LABEL" \
  --label "GREAT_LABEL"
```
