#!/bin/bash

cd "$(dirname "$0")" || exit "$?"
go run update_slack_logs.go update_slack_logs.json slacklog_template/ slacklog_data/ slacklog_pages/
