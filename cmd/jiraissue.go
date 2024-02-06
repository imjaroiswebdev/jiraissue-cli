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
		summary        string
		timeEstimate   string
		description    string
		epicID         string
		labels         *[]string
		components     *[]string
		isDebugEnabled bool
	)

	var rootCmd = &cobra.Command{
		Use:   "jiraissue",
		Short: "Jira CLI to create issues",
		Run: func(cmd *cobra.Command, args []string) {
			assigneeID, ok := os.LookupEnv("JIRA_ASSIGNEE_ID")
			if !ok {
				log.Fatal("Environment Variable JIRA_ASSIGNEE_ID not set")
			}

			// [ ] TODO: Turn this hardcoded values into env var or params
			issueType := "Story"
			priorityID := "2"

			issueKey, err := jiraissue.CreateJiraIssue(summary, timeEstimate, description, epicID, issueType, priorityID, assigneeID, *components, *labels, isDebugEnabled)
			if err != nil {
				log.Fatal(err)
			}

			var successMsg string
			jirasubdomain := os.Getenv("JIRA_SUBDOMAIN")
			if jirasubdomain == "" {
				successMsg = fmt.Sprintf("Issue created. %s", issueKey)
			}
			successMsg = fmt.Sprintf("Issue created. Link to issue https://%s.atlassian.net/browse/%s", jirasubdomain, issueKey)
			fmt.Println(successMsg)
		},
	}

	rootCmd.Flags().StringVarP(&summary, "summary", "s", "", "Summary of the issue (required)")
	rootCmd.Flags().StringVarP(&timeEstimate, "time", "t", "", "Time estimation (required)")
	rootCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the issue")
	rootCmd.Flags().StringVarP(&epicID, "epic", "e", "", "Epic ID")
	rootCmd.Flags().BoolVar(&isDebugEnabled, "debug", false, "Enable debug of API calls")
	components = rootCmd.Flags().StringArrayP("component", "c", []string{}, "Components names separated list")
	labels = rootCmd.Flags().StringArrayP("label", "l", []string{}, "Labels names separated list")

	rootCmd.MarkFlagRequired("summary")
	rootCmd.MarkFlagRequired("time")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
