# Jira issue creator cli

## Description

This tool is a CLI application for creating issues in Jira based on a specific workflow a follow to manage Jira issues, at this point it is not intended to be of general use for the majority of cases.

Supports creation of single Jira Issues through direct use `jiraissue` command passing in each Issue field through command flags. In addition, also support bulk creation of Jira Issues utilizing a **csv** file, which its path can be passed in with the `csv` flag.

## Installation
1. Clone the repository.
2. Build the application: `go build`

## Configuration
Set the following environment variables:
- `JIRA_API_TOKEN`: Your Jira API token. [(How to get one)](https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/)
- `JIRA_PROJECT_KEY`: The key of your Jira project. (e.g., **PROJ**)
- `JIRA_SUBDOMAIN`: (Optional) Your Jira sub domain. (e.g., `https://<subdomain>.atlassian.net`)
- `JIRA_ASSIGNEE_ID`: Jira issue assignee id. (e.g., `62703a40ca1fae106ae98fed`)

## Usage
Run the application with the necessary parameters:

```bash
$ go run ./cmd --help
Jira CLI to create issues

Usage:
  jiraissue [flags]

Flags:
      --assignee string         Issue assignee id
  -c, --component stringArray   Components names separated list
      --csv string              CSV file path for bulk Jira issues creation (e.g., ./jira_issues.csv)
      --debug                   Enable debug of API calls
  -d, --description string      Description of the issue
  -e, --epic string             Epic Key (e.g., PROJ-2948)
      --fixversion string       Fix version name for the issue
  -h, --help                    help for jiraissue
  -l, --label stringArray       Labels names separated list
  -s, --summary string          Summary of the issue (required for single Issue creation)
  -t, --time string             Time estimation in hours (e.g., 2h)
```

## Single Jira Issue creation

```sh
# Using JIRA_ASSIGNEE_ID env var
$ go run ./cmd \
  --summary "New Awesome created through jiraissue cli" \
  --time "4h" \
  --description 'Issue created while testing `jiraissue` cli App' \
  --epic "PROJ-1758" \
  --component "BACKEND" \
  --component "MIDDLEWARE" \
  --component "FRONTEND" \
  --label "AWESOME_LABEL" \
  --label "GREAT_LABEL"
# output
Issue created. Link to issue https://pagerduty.atlassian.net/browse/PROJ-2920
```

## Bulk creation of Jira issues using CSV file

### Content of CSV file `jira_issues.csv` used in the example

```csv
summary;description;time;epic;components;labels;fixVersionName
First issue created with CSV;This is a safe to delete issue created while testing jiraissue cli App;2h;PROJ-2496;BACKEND, MIDDLEWARE;AWESOME_LABEL, GREAT_LABEL;August 2024
Second issue created with CSV;This is a safe to delete issue created while testing jiraissue cli App;2h;PROJ-2496;FRONTEND;GREAT_LABEL;August 2024
```

```sh
# Using --assignee flag
$ go run ./cmd \
  --assignee '62703a40ca1fef006ae18fed' \
  --csv ./jira_issues.csv
# output
Issue created. Link to issue https://pagerduty.atlassian.net/browse/PROJ-2926
Issue created. Link to issue https://pagerduty.atlassian.net/browse/PROJ-2927
```

## Using a `.env` file

Create a `.env` file in the root of the project with the following content:

```sh
JIRA_API_TOKEN=<your_jira_api_token>
JIRA_PROJECT_KEY=<your_jira_project_key>
JIRA_SUBDOMAIN=<your_jira_subdomain>
```

Then a way to source the `.env` file variables is to run the following command:

```sh
$ export $(cat .env | xargs)
```

You should be ready to run the application with the necessary parameters.
