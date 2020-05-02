#!/bin/bash

set -eu

cd "$(dirname "$0")" || exit "$?"
go run ./main.go download-emoji ../emojis/ ../slacklog_data/emoji.json
