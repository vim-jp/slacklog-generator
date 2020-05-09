#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

go run . download-files slacklog_data/ files/
