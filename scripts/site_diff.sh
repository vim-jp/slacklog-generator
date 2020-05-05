#!/bin/bash

set -eu

force=0
clean=0
update=0
docker=0

while getopts fcud OPT ; do
  case $OPT in
    f) force=1 ;;
    c) clean=1 ;;
    u) update=1 ;;
    d) docker=1 ;;
  esac
done

cd "$(dirname "$0")/.." || exit "$?"

outrootdir=tmp/site_diff
current_pages=${outrootdir}/current
cmd=${outrootdir}/slacklog-tools

build_tool() {
  cd scripts
  go build -o ../${cmd} ./main.go
  cd ..
}

# generate-html サブコマンドとjekyll buildを実行して指定ディレクトリに出力する
generate_site() {
  id=$1 ; shift
  outdir=${outrootdir}/${id}
  echo "jekyll build to: ${outdir}" 1>&2
  rm -rf slacklog_pages
  build_tool
  ${cmd} generate-html scripts/config.json slacklog_template/ slacklog_data/ slacklog_pages/ > ${outdir}.generate-html.log 2>&1
  rm -f ${cmd}
  rm -rf ${outdir}
  if [ $docker -ne 0 ] ; then
    docker run --rm -it --volume="$PWD:/srv/jekyll" jekyll/jekyll:pages jekyll build -d ${outdir} > ${outdir}.docker-jekyll-build.log 2>&1
  else
    jekyll build -d ${outdir} > ${outdir}.jekyll-build.log 2>&1
  fi
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
  echo "move to base: $(git rev-parse HEAD)" 2>&1
  generate_site ${base_commit}

  # 退避したHEADと変更を復帰する
  git reset -q --hard ${current_commit}
  if [ $has_changes -ne 0 ] ; then
    git stash pop -q
  fi

  echo "return to current: $(git rev-parse HEAD)" 2>&1
fi

# 差分を出力
diff -uNr -x sitemap.xml ${base_pages} ${current_pages}
