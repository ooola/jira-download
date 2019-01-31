package main

import (
	"flag"
	"fmt"
	"github.com/andlabs/ui"
	jira "github.com/andygrunwald/go-jira"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var jiraClient *jira.Client

var numUserComments = make(map[string]int)
var lock sync.Mutex

var usersChan = make(chan string)

var box *ui.Box

var (
	username       = flag.String("username", "", "Jira Username")
	password       = flag.String("password", "", "Jira Auth Key")
	jiraURL        = flag.String("jiraURL", "https://optimizely.atlassian.net", "URL where Jira is running")
	jiraQuery      = flag.String("jiraQuery", "labels IN (soc2_IRL_fy19)", "The jira query of which ticket attachments to fetch")
	searchComments = flag.Bool("searchComments", false, "Searches comments in tickets for number of authors and comment count")
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
		lock.Lock()
		numUserComments[comment.Author.Name]++
		lock.Unlock()
	}
}

// callback function provided to jira.SearchPages
func handleIssue(issue jira.Issue) error {
	if *searchComments {
		printCommentAuthors(jiraClient, issue.Key)
	} else {
		downloadIssueAttachments(jiraClient, issue.Key)
	}
	return nil
}

//funcUpdateUI()

func main() {
	flag.Parse()
	*searchComments = true

	err2 := ui.Main(func() {
		userInput := ui.NewEntry()
		passInput := ui.NewEntry()
		queryInput := ui.NewEntry()
		button := ui.NewButton("Execute Query")
		greeting := ui.NewLabel("")
		box = ui.NewVerticalBox()
		box.Append(ui.NewLabel("Enter your username:"), false)
		box.Append(userInput, false)
		box.Append(ui.NewLabel("Enter password (token):"), false)
		box.Append(passInput, false)
		box.Append(ui.NewLabel("Enter query:"), false)
		box.Append(queryInput, false)
		box.Append(button, false)
		box.Append(greeting, false)
		window := ui.NewWindow("Jira Query Tool", 200, 100, false)
		window.SetMargined(true)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {
			greeting.SetText("Executing query, " + queryInput.Text() + " ...")
			tp := jira.BasicAuthTransport{
				Username: userInput.Text(),
				Password: passInput.Text(),
			}
			client, err := jira.NewClient(tp.Client(), strings.TrimSpace(*jiraURL))
			if err != nil {
				//greeting.SetText("\nerror: " + string(err))
				return
			}
			jiraClient = client

			if err != nil {
				//greeting.SetText("\nerror: %v\n", err.Error())
				return
			}

			err = client.Issue.SearchPages(queryInput.Text(), nil, handleIssue)
			if err != nil {
				greeting.SetText(err.Error())
				return
			}
			if *searchComments {
				for name, count := range numUserComments {
					c := strconv.Itoa(count)
					line := name + " " + c
					//fmt.Printf("%s %d\n", name, count)
					box.Append(ui.NewLabel(line), false)
				}
			}
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err2 != nil {
		panic(err2)
	}
}
