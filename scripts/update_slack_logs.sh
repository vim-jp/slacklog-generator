#!/bin/bash

cd "$(dirname "$0")/.." || exit "$?"
go run scripts/update_slack_logs.go scripts/update_slack_logs.json slacklog_template/ slacklog_data/ slacklog_pages/
