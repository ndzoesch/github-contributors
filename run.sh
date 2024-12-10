#!/bin/bash

go run main.go > all_commits.json &&\
jq -r '.[].commit.author.name' all_commits.json | sort | uniq > authors.csv