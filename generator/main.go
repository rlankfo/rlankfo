package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v40/github"
)

var (
	README = `
![Photo of Robbie Lankford](https://github.com/rlankfo/rlankfo/blob/main/20210812_183004_Robbie_Lankford-Medium.jpg?raw=true)

### Hi there 👋 I'm Robbie
- 🔭 I’m currently working on the Grafana Agent
- ⚡ Fun fact: The photo above was taken at Lake Powell Resort before a white water rafting trip in the Grand Canyon.
`
	reviewMyCode = `
#### Review some of my code :eyes:
`

	commentOnAnIssue = `
#### Comment on some issues :fire:
`

	dailyFortune = `
#### Daily Fortune :crescent_moon:
`

	weatherReport = `
#### Weather Report :partly_sunny:
![Weather for Rogers, AR](https://wttr.in/Rogers,%20AR_nuFqp_background=0d1117.png)
`
)

func generateContent() (string, error) {
	result := README
	timestamp := fmt.Sprintf("<sub>README.md generated at %s :trollface:</sub>", time.Now().UTC().String())
	prList, err := getPullRequests()
	if err != nil {
		return "", err
	}
	if prList != "" {
		result = fmt.Sprintf("%s%s%s", result, reviewMyCode, prList)
	}
	issues, err := getIssues()
	if err != nil {
		return "", err
	}
	if issues != "" {
		result = fmt.Sprintf("%s%s%s", result, commentOnAnIssue, issues)
	}
	if fortune, ok := os.LookupEnv("FORTUNE"); ok && fortune != "" {
		result = fmt.Sprintf("%s%s\n```\n%s\n```\n", result, dailyFortune, fortune)
	}
	return fmt.Sprintf("%s%s\n%s\n", result, weatherReport, timestamp), nil
}

func getIssues() (string, error) {
	return searchIssues("is:open is:issue author:rlankfo archived:false", "issues",":call_me_hand:")
}

func getPullRequests() (string, error) {
	return searchIssues("is:open is:pr author:rlankfo archived:false", "pull",":metal:")
}

func searchIssues(query string, separator string, emoji string) (string, error) {
	var (
		ctx    = context.Background()
		gh     = github.NewClient(nil)
		next   = 1
		b      bytes.Buffer
		result *github.IssuesSearchResult
		resp   *github.Response
		err    error
	)

	for next != 0 {
		result, resp, err = gh.Search.Issues(ctx, query, &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page: next,
			},
		})
		if err != nil {
			return "", err
		}
		next = resp.NextPage
		for _, issue := range result.Issues {
			htmlURL := strings.Split(issue.GetHTMLURL(), fmt.Sprintf("/%s/", separator))[0]
			repoName := strings.Split(htmlURL, "https://github.com/")[1]
			b.WriteString(fmt.Sprintf("* %s [%s](%s): [%s](%s)\n",
				emoji, repoName, htmlURL, issue.GetTitle(), issue.GetHTMLURL()))
		}
	}

	return b.String(), nil
}

func main() {
	// generate README content
	content, err := generateContent()
	if err != nil {
		log.Fatalf("failed to generate README.md content: %s", err.Error())
	}

	// write content to README
	f, err := os.OpenFile("README.md", os.O_RDWR, 600)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("failed to close README.md: %s", err.Error())
		}
	}(f)
	if err != nil {
		log.Fatalf("failed to open README.md: %s", err.Error())
	}
	err = f.Truncate(0)
	if err != nil {
		log.Fatalf("failed to truncate file: %s", err.Error())
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatalf("seek error: %s", err.Error())
	}
	_, err = fmt.Fprintf(f, "%s", content)
	if err != nil {
		log.Fatalf("error writing to file: %s", err.Error())
	}
}
