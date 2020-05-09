#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

go run . download-emoji _logdata/emojis/ _logdata/slacklog_data/emoji.json
