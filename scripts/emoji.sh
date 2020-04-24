#!/bin/bash

set -eu

cd "$(dirname "$0")" || exit "$?"
go run ./main.go emoji ../emojis/ ../_data/emoji.json
