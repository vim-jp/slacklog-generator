#!/bin/bash

set -eu

outdir="_site"

while getopts o: OPT ; do
  case $OPT in
    o) outdir="$OPTARG" ;;
  esac
done

mkdir -p ${outdir}

cp -a assets ${outdir}

for d in emojis files ; do
  if [ -d _logdata/${d} ] ; then
    cp -a _logdata/${d} ${outdir}
  else
    cp -a ${d} ${outdir}
  fi
done
