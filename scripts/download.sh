#!/bin/bash

set -eu

cd "$(dirname "$0")" || exit "$?"
go run ./main.go download ../slacklog_data/ ../files/
