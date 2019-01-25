package main

import (
	"flag"
	"fmt"
	jira "github.com/andygrunwald/go-jira"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var jiraClient *jira.Client

//var numUserComments map[string]int
var numUserComments = make(map[string]int)

//var usersChan chan string
var usersChan = make(chan string)

var (
	username  = flag.String("username", "", "Jira Username")
	password  = flag.String("password", "", "Jira Auth Key")
	jiraURL   = flag.String("jiraURL", "https://optimizely.atlassian.net", "URL where Jira is running")
	jiraQuery = flag.String("jiraQuery", "labels IN (soc2_IRL_fy19)", "The jira query of which ticket attachments to fetch")
)

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

func downloadAttachment(client *jira.Client, id string, filename string) {
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

	for _, attachment := range issue.Fields.Attachments {
		_, file := filepath.Split(attachment.Content)
		ext := strings.ToLower(filepath.Ext(attachment.Content)) // uppercase file extensions are lame
		if HasImageExt(ext) {
			basename := file[0 : len(file)-len(filepath.Ext(file))]
			tmpfile, err := ioutil.TempFile(".", basename+".*"+ext)
			if err != nil {
				fmt.Printf("Failed to create TempFile\n")
				return
			}
			tmpfile.Close()
			name := tmpfile.Name()
			fmt.Printf("%s --> %s\n", attachment.Content, name)
			go downloadAttachment(client, attachment.ID, name)
		} else {
			fmt.Printf("Skipping %s because it's not an image\n", attachment.Content)
		}
	}
}

func printCommentAuthors(client *jira.Client, jiraKey string) {
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

	for _, comment := range issue.Fields.Comments.Comments {
		//	fmt.Printf("comment: %v\n", comment)
		fmt.Printf("comment author: %v\n", comment.Author.Name)
		// TODO Add count of comments from each author
		//numComments[comment.Author.Name]++
		usersChan <- comment.Author.Name
		//fmt.Printf("sent into channel\n")
		/*
			_, file := filepath.Split(attachment.Content)
			ext := strings.ToLower(filepath.Ext(attachment.Content)) // uppercase file extensions are lame
			if HasImageExt(ext) {
				basename := file[0 : len(file)-len(filepath.Ext(file))]
				tmpfile, err := ioutil.TempFile(".", basename+".*"+ext)
				if err != nil {
					fmt.Printf("Failed to create TempFile\n")
					return
				}
				tmpfile.Close()
				name := tmpfile.Name()
				go downloadAttachment(client, attachment.ID, name)
			} else {
				fmt.Printf("Skipping %s because it's not an image\n", attachment.Content)
			}
		*/
	}
}

// callback function provided to jira.SearchPages
func handleIssue(issue jira.Issue) error {
	//downloadIssueAttachments(jiraClient, issue.Key)
	printCommentAuthors(jiraClient, issue.Key)
	return nil
}

// updates the numUserComments table
func updateUserCommentsTable(users chan string) {

	// read names off the users channel and update the number of comments per user table
	for {
		fmt.Printf("reading from users channel\n")
		select {
		case u := <-usersChan:
			fmt.Printf("got user: %v\n", u)
			numUserComments[u]++
			fmt.Printf("new map: %v\n", numUserComments)
		}
	}

}

func main() {
	flag.Parse()
	//numUserComments = make(map[string]int)
	//usersChan := make(chan string, 20)
	go updateUserCommentsTable(usersChan)

	if len(*username) == 0 || len(*password) == 0 {
		fmt.Printf("No username or API key provided\n")
		os.Exit(1)
	}

	tp := jira.BasicAuthTransport{
		Username: *username,
		Password: *password,
	}
	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(*jiraURL))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	jiraClient = client

	if err != nil {
		fmt.Printf("\nerror: %v\n", err.Error())
		return
	}

	err = client.Issue.SearchPages(*jiraQuery, nil, handleIssue)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	// TODO: must wait for channel to be drained
	fmt.Printf("---%v\n", numUserComments)
	os.Exit(0)
}
