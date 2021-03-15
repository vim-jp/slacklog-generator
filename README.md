# slacklog

## What

A project to htmlize vim-jp Slack logs.

This solves the following problems due to using free tier

* No old messages to see
* Can't see unless you join the workspace, even though you'd like to refer from something such as your blogs
* We lose our knowledge base

## How to contribute

[@tyru](https://twitter.com/_tyru_) and [@thinca](https://twitter.com/thinca) will invite you  to slacklog Team if you contact us via vim-jp Slack or Twitter. We'll share Slack token as well.

How to join vim-jp Slack (Japanese):
<https://vim-jp.org/docs/chat.html>

## What you need to develop

- Go
- (Optional) GNU Make

## Env vars

Create `.env` copiying `.env.template`.
See the details for each env vars in the file.

## How to develop

### Generate HTML

Unarchive logs

```console
$ make logdata
```

Generate HTML

The following commands will generate them under `_site` dir.

```console
scripts/generate_html.sh
scripts/build.sh
```

Or simply run `make` or `gmake`

### Download attached files and emojis

```console
scripts/download_emoji.sh
scripts/download_files.sh
```

### Run dev server

Use your favourite server under `_site`

e.g.

```console
python -m http.server --directory=_site
```

### How to check the diff from geneate-html subcommands

The generate-html output diff from your changes can be checked with this:

```console
$ ./scripts/site_diff.sh
```

TODO translate the following

> `site_diff.sh` では現在のHEADでの generate-html の結果と merge-base での
> geneate-html の結果の diff を取得しています。
> 出力先は `./tmp/site_diff/current/` および
> `./tmp/site_diff/{merge-base-commit-id}/` ディレクトリとなっています。
> 
> merge-base の算出基準はローカルの origin/master です。そのため origin/master が
> リモート(GitHub)の物よりも古いと出力内容が異なり、差分も異なる場合があります。
> `-u` オプションを使うと merge-base の算出前にローカルの origin/master を更新し
> ます。変更がなくても更新にそれなりに時間がかかるため、デフォルトではオフになっ
> ており明示的に指定するようにしています。
> 
> merge-base の出力結果はキャッシュし再利用しています。このキャッシュを無視して強
> 制的に再出力するには `-f` オプションを使ってください。
> 
> ```console
> $ ./scripts/site_diff.sh -f
> ```
> 
> 全てのキャッシュを破棄したい場合には `-c` オプションを使ってください。`-c` オプ
> ションでは `./tmp/site_diff/` ディレクトリを消すだけで差分の出力は行いません。
> 
> ```console
> $ ./scripts/site_diff.sh -c
> ```
> 
> 差分だけを特定のファイルに出力するには `-o {filename}` オプションを使ってくださ
> い。リダイレクト (` > filename`) では差分以外の動作ログも含まれる場合がありま
> す。
> 
> 注意事項: `./scripts/site_diff.sh` は未コミットな変更を stash を用いて保存・復
帰しているため staged な変更が unstaged に巻き戻ることに留意してください。

## How to update log-data

TODO translate the following

> log-data ブランチにはSlackからエクスポートしたデータを格納し、それを本番の生成
> に利用しています。log-data ブランチの更新手順は以下の通りです。
> 
> 1. Slack からログをエクスポート(今はできる人が限られてる)
> 2. ログをワーキングスペースに展開する
> 3. `convert-exported-logs` サブコマンドを実行する
> 
>     ```console
>     $ go run . convert-exported-logs {indir} {outdir}
>     ```
> 
> 4. 更新内容を log-data ブランチに `commit --amend` して `push -f`

## How to see the changes at Pull Request

TODO translate the following

> 以下の手順で Pull Request への `site_diff.sh` の実行結果を
> Artifacts として Web から取得できます。レビューの際に利用してください。
> 
> 1. Pull Request の <b>Checks</b> タブを開く
> 2. <b>CI</b> ワークフロー(右側)を選択
> 3. <b>Compare Pages and Site</b> ジョブ(右側)を選択
> 4. <b>Artifacts</b> ドロワー(左側)を選択
> 5. `diffs-{数値}` アーティファクトをダウンロード
> 
> 以下のスクリーンショットは、上記の選択個所をマーキングしたものです。
> (SSには3つのアーティファクトが表示されますが、現在は2つになっています)
> 
> ![](https://raw.githubusercontent.com/wiki/vim-jp/slacklog-generator/images/where-are-artifacts.png)
> 
> Artifacts はそれぞれ zip としてダウンロードできます。
> `diffs-*.zip` には `sites_diff.sh` の差分が含まれています。
> `log-*.zip` は動作ログが含まれていますが、こちらはCIの動作デバッグ目的のものです。
> 末尾の数値は [`${{ github.run_id }}`](https://help.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context) 由来です。

### Why we output the diff at Artifacts

TODO translate the following

> Artifacts に差分を出力している主な理由は2つあります。1つ目は、小さな変更でも差
> 分をオンライン上のどこかに出力しないと、レビューの負荷が高すぎてそれを解消した
> かったという動機です。
> 
> 2つ目は、テストデータとして実際のログを使っているため、差分とはいえログの一部の
> コピーが消せない状態で永続化されるのを避けたい、という動機です。vim-jp slackで
> は参加者の「忘れられる権利」を尊重しています。
> 
> 以上の理由から消せる状態でデータ=差分をオンライン上にホストできる GitHub
> Actions の Artifacts を利用しています。

## LICNESE

TODO translate the following

> <a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="クリエイティブ・コモンズ・ライセンス" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />この 作品 は <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">クリエイティブ・コモンズ 表示 4.0 国際 ライセンス</a>の下に提供されています。
