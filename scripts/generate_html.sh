#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

go run . generate-html scripts/config.json templates/ slacklog_data/ _site/
