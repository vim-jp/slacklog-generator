#!/bin/bash

cd "$(dirname "$0")" || exit "$?"
go run ./main.go generate-html ./config.json ../templates/ ../slacklog_data/ ../slacklog_pages/
