#!/bin/bash

set -eu

outdir="_site"

while getopts o: OPT ; do
  case $OPT in
    o) outdir="$OPTARG" ;;
  esac
done

mkdir -p ${outdir}

cp -a static/* ${outdir}

for d in emojis files ; do
  if [ ! -d _logdata/${d} ] ; then
    echo "one of input missing. please run 'make logdata' and retry"
    exit 1
  fi
  cp -a _logdata/${d} ${outdir}
done
