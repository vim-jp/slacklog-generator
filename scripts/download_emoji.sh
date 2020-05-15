#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

if  [ ! -d _logdata/slacklog_data/ ] ; then
  echo "one of input missing. please make assure _logdata/slacklog_data/ directory"
  exit 1
fi

go run . download-emoji
