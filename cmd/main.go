package main

import (
	"fmt"
	"log"
	"os"

	jiraissue "github.com/imjaroiswebdev/jiraissue-cli"

	"github.com/spf13/cobra"
)

func main() {
	var (
		assigneeID     string
		summary        string
		timeEstimate   string
		description    string
		epicKey        string
		csvPath        string
		labels         *[]string
		components     *[]string
		isDebugEnabled bool
		isDryRunning   bool
		fixVersionName string
	)

	var rootCmd = &cobra.Command{
		Use:   "jiraissue",
		Short: "Jira CLI to create issues",
		Run: func(cmd *cobra.Command, args []string) {
			// [ ] TODO: Refactor this to handle env vars validation more nicely
			var assignee string
			if assigneeID != "" {
				assignee = assigneeID
			} else {
				var ok bool
				assignee, ok = os.LookupEnv("JIRA_ASSIGNEE_ID")
				if !ok {
					log.Fatal("Issue Assignee should be passed in either through --assignee flag, or Environment Variable JIRA_ASSIGNEE_ID")
				}
			}

			// [ ] TODO: Turn this hardcoded values into env var or params
			issueType := "Story"
			priorityID := "2"

			err := jiraissue.CreateJiraIssue(summary, timeEstimate, description, epicKey, issueType, priorityID, assignee, fixVersionName, csvPath, *components, *labels, isDebugEnabled, isDryRunning)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&summary, "summary", "s", "", "Summary of the issue (required for single Issue creation)")
	rootCmd.Flags().StringVarP(&timeEstimate, "time", "t", "", "Time estimation in hours (e.g., 2h)")
	rootCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the issue")
	rootCmd.Flags().StringVarP(&epicKey, "epic", "e", "", "Epic Key (e.g., PROJ-2948)")
	rootCmd.Flags().StringVar(&assigneeID, "assignee", "", "Issue assignee id")
	rootCmd.Flags().StringVar(&csvPath, "csv", "", "CSV file path for bulk Jira issues creation (e.g., ./jira_issues.csv)")
	rootCmd.Flags().BoolVar(&isDebugEnabled, "debug", false, "Enable debug of API calls")
	rootCmd.Flags().StringVar(&fixVersionName, "fixversion", "", "Fix version name for the issue")
	rootCmd.Flags().BoolVar(&isDryRunning, "dry-run", false, "Dry run mode, no issue will be created, but the payload will be printed")

	components = rootCmd.Flags().StringArrayP("component", "c", []string{}, "Components names separated list")
	labels = rootCmd.Flags().StringArrayP("label", "l", []string{}, "Labels names separated list")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
