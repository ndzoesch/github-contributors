# GitHub Yearly Contributors

This tool fetches all commit data from a given GitHub repository for a specified year, following pagination automatically, and outputs the results as a single JSON array to `stdout`.

A `run.sh` is provided for streamlining the process by parsing the JSON file through `jq` and putting the result in a CSV file.

If you just want the contributors for the current year for `shopware/showpare`, execute `./run.sh`.

## Manual Usage

### Run with flags:

```shell
go mod tidy && go run main.go -owner=shopware -repo=shopware -year=2024 > all_commits.json
```

By default, the owner is `shopware`, the repo is `shopware`, and the year is set to the current year.

### Once you have `all_commits.json`, you can parse it using `jq` and put it into a csv:
   
```shell
jq -r '.[].author.login' all_commits.json | sort | uniq > authors.csv
```

Adjust the jq filter as needed to extract and process the data you want. By default, the GitHub login of the user is used, which is also the GitHub username.

## Avatar download

If you feed the script a csv file containing the logins from the last step, you can download all images from these users by using this command:

```shell
go mod tidy && go run main.go -csv=authors.csv
```

## Authentication

If the repository is private or rate-limits are an issue, create a .env file containing your personal access token for GitHub:

```ini
GITHUB_TOKEN=1234567890abcdefg...
```

## Notes

- Increase `per_page` if necessary by modifying the query parameters in the code.
- Make sure to have `jq` installed if you want to parse the output directly.
