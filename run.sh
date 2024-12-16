#!/bin/bash

go mod tidy && go run main.go > all_commits.json &&\
jq -r '.[].author.login' all_commits.json | sort | uniq > authors.csv