#!/bin/bash

set -eu

copy() {
  if which rsync > /dev/null && [[ -d $1 ]] ; then
    local dir
    dir=$(realpath $1)
    rsync -avP --link-dest=$dir/ $dir/ $2/$1
  else
    cp -a $1 $2
  fi
}

outdir="_site"

while getopts o: OPT ; do
  case $OPT in
    o) outdir="$OPTARG" ;;
  esac
done

mkdir -p ${outdir}

copy assets ${outdir}
copy emojis ${outdir}
copy favicon.ico ${outdir}
copy files ${outdir}
