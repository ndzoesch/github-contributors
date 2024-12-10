package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

func main() {

	currentYear, _, _ := time.Now().Date()

	year := flag.Int("year", currentYear, "year the commits took place in")
	owner := flag.String("owner", "shopware", "the owner of the repo")
	repo := flag.String("repo", "shopware", "the name of the repo")

	flag.Parse()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?since=%d-01-01T00:00:00Z&per_page=100", *owner, *repo, year)

	var allCommits []map[string]interface{}
	client := &http.Client{}

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		// Add auth if you run in a request limit (403 usually):
		// req.Header.Set("Authorization", "token YOUR_TOKEN")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "non-OK HTTP status: %s\n", resp.Status)
			break
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var commits []map[string]interface{}
		if err := json.Unmarshal(body, &commits); err != nil {
			panic(err)
		}

		allCommits = append(allCommits, commits...)

		linkHeader := resp.Header.Get("Link")
		url = parseNextLink(linkHeader)
	}

	// Write allCommits to stdout, use `> out.json` to put it into a file
	output, err := json.MarshalIndent(allCommits, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}

// parseNextLink finds the next page URL from the Link header, following the pagination
func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	// Regex to extract the next link
	re := regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
	matches := re.FindStringSubmatch(linkHeader)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
