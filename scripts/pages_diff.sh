#!/bin/bash

set -eu

force=0
clean=0

while getopts fc OPT ; do
  case $OPT in
    f) force=1 ;;
    c) clean=1 ;;
  esac
done

cd "$(dirname "$0")" || exit "$?"

# generate-html サブコマンドを実行して指定ディレクトリに出力する
generate_html() {
  outdir=$1 ; shift
  rm -rf $outdir
  echo "generate_html to: $outdir" 1>&2
  mkdir -p $outdir
  go run ./main.go generate-html ./config.json ../slacklog_template/ ../slacklog_data/ ${outdir} > ${outdir}.generate-html.log 2>&1
}

outrootdir=../tmp/pages_diff
current_pages=${outrootdir}/current

if [ $clean -ne 0 ] ; then
  echo "clean up $outrootdir" 1>&2
  rm -rf $outrootdir
  exit 0
fi

echo "catching up origin/master" 1>&2
git fetch -q origin master
base_commit=$(git show-branch --merge-base origin/master HEAD)
base_pages=${outrootdir}/${base_commit}

generate_html ${current_pages}

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
  generate_html ${base_pages}

  # 退避したHEADと変更を復帰する
  git reset -q --hard ${current_commit}
  if [ $has_changes -ne 0 ] ; then
    git stash pop -q
  fi

  echo "return to current: $(git rev-parse HEAD)" 2>&1
fi

# 差分を出力
diff -uNr ${base_pages} ${current_pages}
