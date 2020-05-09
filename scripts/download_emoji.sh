#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

go run . download-emoji emojis/ slacklog_data/emoji.json
