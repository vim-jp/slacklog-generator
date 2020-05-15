#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

if [ ! -d _logdata/slacklog_data/ ] ; then
  echo "one of input missing. please run 'make logdata' and retry"
  exit 1
fi

go run . generate-html
