package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
	//"golang.org/x/crypto/ssh/terminal"
)

func downloadFile(filepath string, url string) {
	out, err := os.Create(filepath)
	if err != nil {
		fmt.Print("Create file: %s failed, error: %s\n", filepath, err.Error())
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Print("downloadFile Get Failed: %s\n", err.Error())
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Print("io.Copy Failed: %s\n", err.Error())
	}
}

func downloadIssueAttachments(client *jira.Client, jiraKey string) {
	if client == nil {
		fmt.Printf("client is nil\n")
		return
	}
	if jiraKey == "" {
		fmt.Printf("jiraKey is empty\n")
		return
	}
	issue, response, err := client.Issue.Get(jiraKey, nil)
	if err != nil {
		fmt.Printf("error fetching: %s :%v\n", jiraKey, err.Error())
		return
	}
	if response == nil {
		fmt.Printf("error fetching: %s response is nil\n", jiraKey)
		return
	}
	if response.StatusCode != 200 {
		fmt.Printf("error fetching: %s response code is not 200, it is: \n", jiraKey, response.StatusCode)
		return
	}
	//fmt.Printf("\n\n issue: %s\n", issue)

	fmt.Printf("len issue.Fields: %v\n", issue.Fields)
	fmt.Printf("\n\t\tdata: %v\n", issue.Fields.Attachments)
	fmt.Printf("\n\t\ttype: %v\n", len(issue.Fields.Attachments))
	for i, attachment := range issue.Fields.Attachments {
		current := time.Now()
		fmt.Printf("attament content[%d]: %s\n\n", i, attachment.Content)
		downloadFile(current.String(), attachment.Content)
	}
}

func main() {
	jiraURL := "https://optimizely.atlassian.net"
	tp := jira.BasicAuthTransport{
		Username: "",
		Password: "",
	}
	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(jiraURL))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	//u, _, err := client.User.Get("ola.nordstrom@optimizely.com")
	//u, _, err := client.User.Get("ola.nordstrom")

	if err != nil {
		fmt.Printf("\nerror: %v\n", err.Error())
		return
	}

	//fmt.Printf("\nEmail: %v\nSuccess!\n", u.EmailAddress)

	//issue, _, _ := client.Issue.Get("GRC-3461", nil)
	//fmt.Printf("issue: %s\n", issue)
	//jql=labels%20IN%20(%22soc2_IRL_fy19%22)

	issues, issuesResponse, err := client.Issue.Search("labels IN (soc2_IRL_fy19)", nil)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	//fmt.Printf("issues: %s\n", issues)
	/*
	   "maxResults": 50,
	   "startAt": 0,
	   "total": 109

	*/
	fmt.Printf("issues length: %d, \n\nresponse: %s\n\n", len(issues), issuesResponse)
	//fmt.Printf("issues[0]: %s\n", issues[0])

	for _, issue := range issues {

		//fmt.Printf("issue: %s\n\n", issue)
		//fmt.Printf("key: %s\n\n", issue.Key)
		downloadIssueAttachments(client, issue.Key)
		os.Exit(0)

	}

}
