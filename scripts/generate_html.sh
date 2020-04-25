#!/bin/bash

cd "$(dirname "$0")" || exit "$?"
go run ./main.go generate_html ./config.json ../slacklog_template/ ../slacklog_data/ ../slacklog_pages/
