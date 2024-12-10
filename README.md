# GitHub Yearly Contributors

This tool fetches all commit data from a given GitHub repository for a specified year, following pagination automatically, and outputs the results as a single JSON array to `stdout`.

A `run.sh` is provided for streamlining the process by parsing the JSON file through `jq` and putting the result in a CSV file.

If you just want the contributors for the current year for `shopware/showpare`, execute `./run.sh`.

## Manual Usage

1. Run with flags:
   ```bash
   go run main.go -owner=shopware -repo=shopware -year=2024 > all_commits.json
   ```

   By default, the owner is `shopware`, the repo is `shopware`, and the year is set to the current year.

2. Once you have `all_commits.json`, you can parse it using `jq` and put it into a csv:
   
   ```bash
   jq -r '.[].author.login' all_commits.json | sort | uniq > authors.csv
   ```

   Adjust the jq filter as needed to extract and process the data you want. By default, the GitHub login of the user is used, which is also the GitHub username.

## Authentication

If the repository is private or rate-limits are an issue, set an `Authorization` header in the code:
```go
req.Header.Set("Authorization", "token YOUR_GITHUB_TOKEN")
```

## Notes

- Increase `per_page` if necessary by modifying the query parameters in the code.
- Make sure to have `jq` installed if you want to parse the output directly.
