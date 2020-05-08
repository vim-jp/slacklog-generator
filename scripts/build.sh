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

if [ $docker -ne 0 ] ; then
  docker run --rm -t --volume="$PWD:/srv/jekyll" jekyll/jekyll:pages jekyll build -d ${outdir}
else
  jekyll build -d ${outdir}
fi
