package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var token string

type User struct {
	AvatarURL string `json:"avatar_url"`
}

func main() {

	_ = godotenv.Load() // ignore, if the token is empty we just ignore
	token = os.Getenv("GITHUB_TOKEN")

	currentYear, _, _ := time.Now().Date()

	year := flag.Int("year", currentYear, "year the commits took place in")
	owner := flag.String("owner", "shopware", "the owner of the repo")
	repo := flag.String("repo", "shopware", "the name of the repo")
	csvFile := flag.String("csv", "", "the file name of the csv file containing logins for which avatars should be downloaded")

	flag.Parse()

	if *csvFile != "" {
		downloadAvatars(*csvFile)
		return
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?since=%d-01-01T00:00:00Z&per_page=100", *owner, *repo, year)

	var allCommits []map[string]interface{}
	client := &http.Client{}

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}

		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

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

func fetchUser(ctx context.Context, client *http.Client, login string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/users/"+login, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func downloadAvatars(csvFile string) {
	f, err := os.Open(csvFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	// Create output dir if needed
	outDir := "avatars"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := &http.Client{}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if len(record) == 0 {
			continue
		}
		login := strings.TrimSpace(record[0])
		if login == "" {
			continue
		}

		user, err := fetchUser(ctx, client, login)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching user %s: %v\n", login, err)
			continue
		}

		if user.AvatarURL == "" {
			fmt.Fprintf(os.Stderr, "No avatar URL for user %s\n", login)
			continue
		}

		if err := downloadAvatar(ctx, client, user.AvatarURL, filepath.Join(outDir, login)); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading avatar for %s: %v\n", login, err)
		}
	}
}

func downloadAvatar(ctx context.Context, client *http.Client, url, baseFilename string) error {
	fmt.Println("downloading avatar for: " + baseFilename)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	// Detect content type
	contentType := resp.Header.Get("Content-Type")
	var ext string
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	case "image/gif":
		ext = ".gif"
	default:
		ext = ".jpg" // fallback if unknown
	}

	filename := baseFilename + ext

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
