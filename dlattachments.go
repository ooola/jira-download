package main

import (
	"fmt"
	jira "github.com/andygrunwald/go-jira"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var jiraClient *jira.Client
var sequenceDownloads bool
var FileNamesChannel chan string

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

func writePNGFilenamesToChannel(c chan string) {
	i := 1
	for {
		s := fmt.Sprintf("%04d.png", i)
		c <- s
		i++
	}
}

func intChannels() {
	c := make(chan (string))
	go writePNGFilenamesToChannel(c)

	for i := 0; i < 10; i++ {
		s := <-c
		fmt.Println("%s", s)
	}
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

	//fmt.Printf("len issue.Fields: %v\n", issue.Fields)
	//fmt.Printf("\n\t\tdata: %v\n", issue.Fields.Attachments)
	//fmt.Printf("\n\t\ttype: %v\n", len(issue.Fields.Attachments))
	for _, attachment := range issue.Fields.Attachments {
		_, file := filepath.Split(attachment.Content)
		ext := strings.ToLower(filepath.Ext(attachment.Content))
		if HasImageExt(ext) {
			basename := file[0 : len(file)-len(filepath.Ext(file))]
			tmpfile, err := ioutil.TempFile(".", basename+".*"+ext)
			if err != nil {
				fmt.Printf("Failed to create TempFile\n")
				return
			}
			tmpfile.Close()
			var name string
			if sequenceDownloads {
				name = <-FileNamesChannel
			} else {
				name = tmpfile.Name()
			}
			fmt.Printf("%s --> %s\n", attachment.Content, name)
			//downloadFile(client, name, attachment.Content)
			go downloadAttachment(client, attachment.ID, name)
		} else {
			fmt.Printf("Skipping %s because it's not an image\n", attachment.Content)
		}
	}
}

func handleIssue(issue jira.Issue) error {
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
	sequenceDownloads = false

	if err != nil {
		fmt.Printf("\nerror: %v\n", err.Error())
		return
	}

	if sequenceDownloads {
		FileNamesChannel := make(chan (string))
		go writePNGFilenamesToChannel(FileNamesChannel)
		fmt.Println("first filename: %s", <-FileNamesChannel)
	}
	err = client.Issue.SearchPages("labels IN (soc2_IRL_fy19)", nil, handleIssue)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
}
