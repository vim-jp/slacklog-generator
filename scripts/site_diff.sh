#!/bin/bash

set -eu

force=0
clean=0
update=0
outdiff=""

while getopts fcuo: OPT ; do
  case $OPT in
    f) force=1 ;;
    c) clean=1 ;;
    u) update=1 ;;
    o) outdiff="$OPTARG" ;;
  esac
done

cd "$(dirname "$0")/.." || exit "$?"

outrootdir=tmp/site_diff
current_pages=${outrootdir}/current
cmd=${outrootdir}/slacklog-tools

build_tool() {
  if [ -f main.go ] ; then
    go build -o ${cmd} . 1>&2
  else
    cd scripts
    go build -o ../${cmd} ./main.go 1>&2
    cd ..
  fi
}

# generate-html サブコマンドとjekyll buildを実行して指定ディレクトリに出力する
generate_site() {
  id=$1 ; shift
  outdir=${outrootdir}/${id}
  echo "jekyll build to: ${outdir}" 1>&2
  rm -rf slacklog_pages
  build_tool
  tmpldir=slacklog_template/
  if [ -d templates ] ; then
    tmpldir=templates/
  fi
  ${cmd} generate-html scripts/config.json ${tmpldir} slacklog_data/ slacklog_pages/ > ${outdir}.generate-html.log 2>&1
  rm -f ${cmd}
  rm -rf ${outdir}
  ./scripts/build.sh -o $outdir > ${outdir}.build.log 2>&1
}

if [ $clean -ne 0 ] ; then
  echo "clean up $outrootdir" 1>&2
  rm -rf $outrootdir
  exit 0
fi

generate_site "current"

if [ $update -ne 0 ] ; then
  echo "catching up origin/master" 1>&2
  git fetch -q origin master
fi

base_commit=$(git show-branch --merge-base origin/master HEAD)
base_pages=${outrootdir}/${base_commit}

if [ $force -ne 0 -o ! \( -d $base_pages \) ] ; then
  echo "base commit: $base_commit" 1>&2

  # 現在の変更とHEADの commit ID を退避する
  has_changes=$(git status -s -uno | wc -l)
  if [ $has_changes -ne 0 ] ; then
    git stash push -q
  fi
  current_commit=$(git rev-parse HEAD)

  # merge-base に巻き戻し generate-html を実行する
  git reset -q --hard ${base_commit}
  echo "move to base: $(git rev-parse HEAD)" 1>&2
  generate_site ${base_commit}

  # 退避したHEADと変更を復帰する
  git reset -q --hard ${current_commit}
  if [ $has_changes -ne 0 ] ; then
    git stash pop -q
  fi

  echo "return to current: $(git rev-parse HEAD)" 1>&2
fi

# 差分を出力
if [ x"$outdiff" = x ] ; then
  echo "" 1>&2
  diff -uNr -x sitemap.xml ${base_pages} ${current_pages} || true
else
  diff -uNr -x sitemap.xml ${base_pages} ${current_pages} > "$outdiff" || true
fi
