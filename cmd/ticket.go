package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type JiraIssue struct {
	Expand string `json:"expand"`
	ID     string `json:"id"`
	Self   string `json:"self"`
	Key    string `json:"key"`
	Fields struct {
		Parent struct {
			ID     string `json:"id"`
			Key    string `json:"key"`
			Self   string `json:"self"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Self           string `json:"self"`
					Description    string `json:"description"`
					IconURL        string `json:"iconUrl"`
					Name           string `json:"name"`
					ID             string `json:"id"`
					StatusCategory struct {
						Self      string `json:"self"`
						ID        int    `json:"id"`
						Key       string `json:"key"`
						ColorName string `json:"colorName"`
						Name      string `json:"name"`
					} `json:"statusCategory"`
				} `json:"status"`
				Priority struct {
					Self    string `json:"self"`
					IconURL string `json:"iconUrl"`
					Name    string `json:"name"`
					ID      string `json:"id"`
				} `json:"priority"`
				Issuetype struct {
					Self           string `json:"self"`
					ID             string `json:"id"`
					Description    string `json:"description"`
					IconURL        string `json:"iconUrl"`
					Name           string `json:"name"`
					Subtask        bool   `json:"subtask"`
					HierarchyLevel int    `json:"hierarchyLevel"`
				} `json:"issuetype"`
			} `json:"fields"`
		} `json:"parent"`
		Resolution struct {
			Self        string `json:"self"`
			ID          string `json:"id"`
			Description string `json:"description"`
			Name        string `json:"name"`
		} `json:"resolution"`
		LastViewed string `json:"lastViewed"`
		Priority   struct {
			Self    string `json:"self"`
			IconURL string `json:"iconUrl"`
			Name    string `json:"name"`
			ID      string `json:"id"`
		} `json:"priority"`
		Labels                        []interface{} `json:"labels"`
		Aggregatetimeoriginalestimate interface{}   `json:"aggregatetimeoriginalestimate"`
		Timeestimate                  interface{}   `json:"timeestimate"`
		Issuelinks                    []interface{} `json:"issuelinks"`
		Assignee                      struct {
			Self         string `json:"self"`
			AccountID    string `json:"accountId"`
			EmailAddress string `json:"emailAddress"`
			DisplayName  string `json:"displayName"`
			Active       bool   `json:"active"`
			TimeZone     string `json:"timeZone"`
			AccountType  string `json:"accountType"`
		} `json:"assignee"`
		Status struct {
			Self           string `json:"self"`
			Description    string `json:"description"`
			IconURL        string `json:"iconUrl"`
			Name           string `json:"name"`
			ID             string `json:"id"`
			StatusCategory struct {
				Self      string `json:"self"`
				ID        int    `json:"id"`
				Key       string `json:"key"`
				ColorName string `json:"colorName"`
				Name      string `json:"name"`
			} `json:"statusCategory"`
		} `json:"status"`
		Components            []interface{} `json:"components"`
		Aggregatetimeestimate interface{}   `json:"aggregatetimeestimate"`
		Creator               struct {
			Self        string `json:"self"`
			AccountID   string `json:"accountId"`
			DisplayName string `json:"displayName"`
			Active      bool   `json:"active"`
			TimeZone    string `json:"timeZone"`
			AccountType string `json:"accountType"`
		} `json:"creator"`
		Subtasks []interface{} `json:"subtasks"`
		Reporter struct {
			Self        string `json:"self"`
			AccountID   string `json:"accountId"`
			DisplayName string `json:"displayName"`
			Active      bool   `json:"active"`
			TimeZone    string `json:"timeZone"`
			AccountType string `json:"accountType"`
		} `json:"reporter"`
		Aggregateprogress struct {
			Progress int `json:"progress"`
			Total    int `json:"total"`
		} `json:"aggregateprogress"`
		Issuetype struct {
			Self           string `json:"self"`
			ID             string `json:"id"`
			Description    string `json:"description"`
			IconURL        string `json:"iconUrl"`
			Name           string `json:"name"`
			Subtask        bool   `json:"subtask"`
			AvatarID       int    `json:"avatarId"`
			HierarchyLevel int    `json:"hierarchyLevel"`
		} `json:"issuetype"`
		Timespent interface{} `json:"timespent"`
		Project   struct {
			Self           string `json:"self"`
			ID             string `json:"id"`
			Key            string `json:"key"`
			Name           string `json:"name"`
			ProjectTypeKey string `json:"projectTypeKey"`
			Simplified     bool   `json:"simplified"`
			AvatarUrls     struct {
				Four8X48  string `json:"48x48"`
				Two4X24   string `json:"24x24"`
				One6X16   string `json:"16x16"`
				Three2X32 string `json:"32x32"`
			} `json:"avatarUrls"`
		} `json:"project"`
		Resolutiondate string `json:"resolutiondate"`
		Watches        struct {
			Self       string `json:"self"`
			WatchCount int    `json:"watchCount"`
			IsWatching bool   `json:"isWatching"`
		} `json:"watches"`
		Created              string      `json:"created"`
		Updated              string      `json:"updated"`
		Timeoriginalestimate interface{} `json:"timeoriginalestimate"`
		Customfield10096     interface{} `json:"customfield_10096"`
		Description          struct {
			Version int    `json:"version"`
			Type    string `json:"type"`
			Content []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"content"`
		} `json:"description"`
		Timetracking struct {
		} `json:"timetracking"`
		Security    interface{}   `json:"security"`
		Attachment  []interface{} `json:"attachment"`
		Summary     string        `json:"summary"`
		Environment interface{}   `json:"environment"`
		Duedate     interface{}   `json:"duedate"`
		Comment     struct {
			Comments   []interface{} `json:"comments"`
			Self       string        `json:"self"`
			MaxResults int           `json:"maxResults"`
			Total      int           `json:"total"`
			StartAt    int           `json:"startAt"`
		} `json:"comment"`
	} `json:"fields"`
}

var ticketCmd = &cobra.Command{
	Use:   "ticket [id]",
	Short: "Get ticket information from JIRA",
	Long:  `This command retrieves ticket information from JIRA.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		jiraApiToken := viper.GetString("jira_api_token")
		jiraEmail := viper.GetString("jira_email")
		ticketId := ""

		if jiraApiToken == "" || jiraEmail == "" {
			fmt.Println("JIRA email and API token are required")
			os.Exit(1)
		}

		if len(args) == 0 {
			fmt.Println("Getting current ticket from branch...")
			ticketId, err := GetTicketFromBranch()
			if err != nil {
				fmt.Println("Error getting ticket from branch: Are you on a ticket branch?", err)
				os.Exit(1)
			}
			fmt.Println("Ticket ID:", ticketId)
			os.Exit(1)
		}

		if ticketId == "" {
			ticketId = args[0]
		}

		url := fmt.Sprintf("https://triipteam.atlassian.net/rest/api/3/issue/%s", ticketId)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			os.Exit(1)
		}

		req.Header.Set("Authorization", "Basic "+basicAuth(jiraEmail, jiraApiToken))
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request to server:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("Received non-200 response: %d\n", resp.StatusCode)
			os.Exit(1)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(bodyBytes))
		var issue JiraIssue
		if err := json.Unmarshal(bodyBytes, &issue); err != nil {
			fmt.Println("Error parsing JSON response:", err)
			os.Exit(1)
		}

		const timeFormat = "2006-01-02T15:04:05.000-0700"

		createdTime, _ := time.Parse(timeFormat, issue.Fields.Created)
		createdTime = createdTime.In(time.FixedZone("CET", 3600))
		humanizedCreatedTime := humanize.Time(createdTime)

		updatedTime, _ := time.Parse(timeFormat, issue.Fields.Updated)
		updatedTime = updatedTime.In(time.FixedZone("CET", 3600))
		humanizedUpdatedTime := humanize.Time(updatedTime)

		resolutionTime, _ := time.Parse(timeFormat, issue.Fields.Updated)
		resolutionTime = resolutionTime.In(time.FixedZone("CET", 3600))
		humanizedResolutionTime := humanize.Time(resolutionTime)

		cyan := color.New(color.FgHiCyan).SprintfFunc()
		blue := color.New(color.FgHiBlue).SprintfFunc()
		urlizeString := func(url string) string {
			return "\033]8;;" + url + "\033\\" + url + "\033]8;;\033\\"
		}

		fmt.Printf("%-25s\t%s\n", cyan("Issue ID:"), blue(issue.Key+" - "+issue.Fields.Summary))
		fmt.Printf("%-25s\t%s\n", cyan("Status:"), blue(issue.Fields.Status.Name))
		fmt.Printf("%-25s\t%s\n", cyan("Assignee:"), blue(issue.Fields.Assignee.DisplayName))
		fmt.Printf("%-25s\t%s\n", cyan("Reporter:"), blue(issue.Fields.Reporter.DisplayName))
		fmt.Printf("%-25s\t%s\n", cyan("Created:"), blue(issue.Fields.Created+" CET "+"("+humanizedCreatedTime+")"))
		fmt.Printf("%-25s\t%s\n", cyan("Updated:"), blue(issue.Fields.Updated+" CET "+"("+humanizedUpdatedTime+")"))
		fmt.Printf("%-25s\t%s\n", cyan("Status:"), blue(issue.Fields.Resolution.Name))
		fmt.Printf("%-25s\t%s\n", cyan("Resolution Date:"), blue(issue.Fields.Resolutiondate+" CET "+"("+humanizedResolutionTime+")"))
		fmt.Printf("%-25s\t%s\n", cyan("URL:"), blue(urlizeString(issue.Self)))

	},
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetTicketFromBranch() (string, error) {
	// Run git command to get the current branch name
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", errors.New("failed to get the current git branch")
	}

	// Get the branch name and trim whitespace
	branch := strings.TrimSpace(out.String())

	// Strip prefixes as specified
	prefixes := []string{"us/", "st/", "hotfix/", "bug/", "fix/"}
	for _, prefix := range prefixes {
		branch = strings.TrimPrefix(branch, prefix)
	}

	// Check for specified starting strings
	if strings.HasPrefix(branch, "SCRUM-") || strings.HasPrefix(branch, "BUG-") {
		return strings.TrimSpace(branch), nil
	}

	// If none of the conditions are met, return an error
	return "", errors.New("branch name does not start with 'SCRUM-' or 'BUG-'")
}

func init() {
	rootCmd.AddCommand(ticketCmd)
}
