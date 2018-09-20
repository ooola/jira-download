package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
	//"golang.org/x/crypto/ssh/terminal"
)

var jiraClient *jira.Client

// returns true if the filename has an image extension
func HasImageExt(file string) bool {
	e := filepath.Ext(file)
	if strings.EqualFold(e, ".png") ||
		strings.EqualFold(e, ".gif") ||
		strings.EqualFold(e, ".jpg") ||
		strings.EqualFold(e, ".pdf") ||
		strings.EqualFold(e, ".bpm") ||
		strings.EqualFold(e, ".tiff") ||
		strings.EqualFold(e, ".svg") {
		return true
	}
	return false
}

func downloadFile(url string, filename string) {
	fmt.Printf("downloading %s --> %s\n", url, filename)
	out, err := os.Create(filename)
	if err != nil {
		fmt.Print("Create file: %s failed, error: %s\n", filename, err.Error())
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

func downloadAttachment(client *jira.Client, id string, filename string) {
	fmt.Printf("downloading attachment %s --> %s\n", id, filename)
	out, err := os.Create(filename)
	if err != nil {
		fmt.Print("Create file: %s failed, error: %s\n", filename, err.Error())
	}
	defer out.Close()

	resp, err := client.Issue.DownloadAttachment(id)
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
		_, file := filepath.Split(attachment.Content)
		timeString := strings.Replace(current.String(), " ", "-", -1)
		name := timeString + "-" + file
		fmt.Printf("timeString: %s, name: %s, file: %s\n", timeString, name, file)
		//downloadFile(client, name, attachment.Content)
		downloadAttachment(client, attachment.ID, file)
		os.Exit(0)
	}
}

func handleIssue(issue jira.Issue) error {
	fmt.Printf("Handling %v\n", issue)
	downloadIssueAttachments(jiraClient, issue.Key)
	return nil
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
	jiraClient = client

	if err != nil {
		fmt.Printf("\nerror: %v\n", err.Error())
		return
	}

	err = client.Issue.SearchPages("labels IN (soc2_IRL_fy19)", nil, handleIssue)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
}
