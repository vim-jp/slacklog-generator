#!/bin/bash

set -eu

cd "$(dirname "$0")/.." || exit "$?"

CMD=${1:-serve}
shift || true

if which make > /dev/null 2>&1; then
  make
else
  if [[ ! -d slacklog_data ]]; then
    git archive origin/log-data | tar x
  fi
  ./scripts/generate_html.sh
fi

docker run --rm -it --volume="$PWD:/srv/jekyll" -p "4000:4000" jekyll/jekyll:pages jekyll ${CMD} "$@"
