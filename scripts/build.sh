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
cp -a emojis ${outdir}
cp -p favicon.ico ${outdir}
cp -a files ${outdir}
