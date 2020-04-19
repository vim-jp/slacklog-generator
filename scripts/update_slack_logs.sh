#!/bin/bash

cd "$(dirname "$0")/.." || exit "$?"
go run scripts/main.go scripts/update_slack_logs.json slacklog_template/ slacklog_data/ slacklog_pages/
