#!/bin/bash

set -eu

docker=0
outdir="_site"

while getopts d:o: OPT ; do
  case $OPT in
    d) docker="$OPTARG" ;;
    o) outdir="$OPTARG" ;;
  esac
done

cd "$(dirname "$0")/.." || exit "$?"
make slacklog_pages
mkdir -p ${outdir}
cp -a assets ${outdir}
cp -a emojis ${outdir}
cp favicon.ico ${outdir}
cp -a files ${outdir}
cp -a slacklog_pages/* ${outdir}
