# slacklog

## これは何

vim-jp Slack のログを HTML 化するプロジェクトです。

Slack 上では無料枠のため

* 古いメッセージが見れない
* ワークスペースに参加していない人には見えないため、ブログ等から引用などして参照したい
* 知見が消えるのはもったいない

といった問題があり、それらを解決するため作られました。

## 開発に参加するには

興味を持った方は vim-jp Slack や [@tyru](https://twitter.com/_tyru_), [@thinca](https://twitter.com/thinca) 等に声をかけて頂ければ GitHub の slacklog Team に招待します (Slack token などもその際共有します)。<br>
vim-jp Slack への参加方法はこちらをどうぞ。<br>
[vim-jp » vim-jpのチャットルームについて](https://vim-jp.org/docs/chat.html)

## 開発に必要なもの

- Go
- ローカルで開発する場合 (Docker を使う場合は不要)
  - Ruby
  - Jekyll
  - (あれば)GNU Make

## 環境変数

`scripts/.env.template` からコピーして `scripts/.env` ファイルを作成してください。
(各環境変数の説明はファイルを参照)

## 開発方法

### Docker を使う場合

```console
./scripts/docker.sh
```

### ローカルで開発する場合

#### HTML の生成

ログを展開

```console
curl -Ls https://github.com/vim-jp/slacklog/archive/log-data.tar.gz | tar xz --strip-components=1 --exclude=.github
```

Jekyll に必要な HTML を生成

```console
scripts/generate_html.sh
```

GNU Makeがあれば`make`もしくは`gmake`を実行するだけで生成されます

#### 添付ファイルと絵文字のダウンロード

```console
scripts/download_emoji.sh
scripts/download_files.sh
```

#### 開発サーバーの起動

Jekyll のインストール(初回のみ)

```console
bundle install
```

開発サーバーの起動

```console
bundle exec jekyll serve -w
```

#### `jekyll build` の出力の差分の確認方法

以下のコマンドで自分が変更した結果として生じた `jekyll build` の出力内容の差分
を確認できます。

```console
$ ./scripts/site_diff.sh
```

`site_diff.sh` では現在のHEADでの `jekyll build` の結果と
merge-base での `jekyll build` の結果の diff を取得しています。
出力先は `./tmp/site_diff/current/` および
`./tmp/site_diff/{merge-base-commit-id}/` ディレクトリとなっています。

デフォルトではローカルにインストールした jekyll を使います。dockerのjekyllを使
用する場合には `-d` オプションを指定してください。

merge-base の算出基準はローカルの origin/master です。そのため origin/master が
リモート(GitHub)の物よりも古いと出力内容が異なり、差分も異なる場合があります。
`-u` オプションを使うと merge-base の算出前にローカルの origin/master を更新し
ます。変更がなくても更新にそれなりに時間がかかるため、デフォルトではオフになっ
ており明示的に指定するようにしています。

merge-base の出力結果はキャッシュし再利用しています。このキャッシュを無視して強
制的に再出力するには `-f` オプションを使ってください。

```console
$ ./scripts/site_diff.sh -f
```

全てのキャッシュを破棄したい場合には `-c` オプションを使ってください。`-c` オプ
ションでは `./tmp/site_diff/` ディレクトリを消すだけで差分の出力は行いません。

```console
$ ./scripts/site_diff.sh -c
```

差分だけを特定のファイルに出力するには `-o {filename}` オプションを使ってくださ
い。リダイレクト (` > filename`) では差分以外の動作ログも含まれる場合がありま
す。

注意事項: `./scripts/site_diff.sh` は未コミットな変更を stash を用いて保存・復
帰しているため staged な変更が unstaged に巻き戻ることに留意してください。

#### geneate-html サブコマンドの出力の差分の確認方法

以下のコマンドで自分が変更した結果として生じた generate-html の出力内容の差分
を確認できます。

```console
$ ./scripts/pages_diff.sh
```

`pages_diff.sh` では現在のHEADでの generate-html の結果と merge-base での
geneate-html の結果の diff を取得しています。
出力先は `./tmp/pages_diff/current/` および
`./tmp/pages_diff/{merge-base-commit-id}/` ディレクトリとなっています。

merge-base の算出基準はローカルの origin/master です。そのため origin/master が
リモート(GitHub)の物よりも古いと出力内容が異なり、差分も異なる場合があります。
`-u` オプションを使うと merge-base の算出前にローカルの origin/master を更新し
ます。変更がなくても更新にそれなりに時間がかかるため、デフォルトではオフになっ
ており明示的に指定するようにしています。

merge-base の出力結果はキャッシュし再利用しています。このキャッシュを無視して強
制的に再出力するには `-f` オプションを使ってください。

```console
$ ./scripts/pages_diff.sh -f
```

全てのキャッシュを破棄したい場合には `-c` オプションを使ってください。`-c` オプ
ションでは `./tmp/pages_diff/` ディレクトリを消すだけで差分の出力は行いません。

```console
$ ./scripts/pages_diff.sh -c
```

差分だけを特定のファイルに出力するには `-o {filename}` オプションを使ってくださ
い。リダイレクト (` > filename`) では差分以外の動作ログも含まれる場合がありま
す。

注意事項: `./scripts/pages_diff.sh` は未コミットな変更を stash を用いて保存・復
帰しているため staged な変更が unstaged に巻き戻ることに留意してください。

## log-data の更新手順

log-data ブランチにはSlackからエクスポートしたデータを格納し、それを本番の生成
に利用しています。log-data ブランチの更新手順は以下の通りです。

1. Slack からログをエクスポート(今はできる人が限られてる)
2. ログをワーキングスペースに展開する
3. `convert-exported-logs` サブコマンドを実行する

    ```console
    $ cd scripts && go run ./main.go convert-exported-logs {indir} {outdir}
    ```

4. 更新内容を log-data ブランチに `commit --amend` して `push -f`

## Pull Request の影響の確認の方法

以下の手順で Pull Request への `pages_diff.sh` と `site_diff.sh` の実行結果を
Artifacts として Web から取得できます。レビューの際に利用してください。

1. Pull Request の <b>Checks</b> タブを開く
2. <b>CI</b> ワークフロー(右側)を選択
3. <b>Compare Pages and Site</b> ジョブ(右側)を選択
4. <b>Artifacts</b> ドロワー(左側)を選択
5. `diffs-{数値}` アーティファクトをダウンロード

以下のスクリーンショットは、上記の選択個所をマーキングしたものです。

![](https://raw.githubusercontent.com/wiki/vim-jp/slacklog-generator/images/where-are-artifacts.png)

Artifacts はそれぞれ zip としてダウンロードできます。
`diffs-*.zip` には `pages_diff.sh` と `site_diff.sh` の両方の差分が含まれています。
`log-*.zip` はそれぞれの動作ログが含まれていますが、こちらはCIの動作デバッグ目的のものです。
末尾の数値は [`${{ github.run_id }}`](https://help.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context) 由来です。

### Artifacts に差分を出力している理由

Artifacts に差分を出力している主な理由は2つあります。1つ目は、小さな変更でも差
分をオンライン上のどこかに出力しないと、レビューの負荷が高すぎてそれを解消した
かったという動機です。

2つ目は、テストデータとして実際のログを使っているため、差分とはいえログの一部の
コピーが消せない状態で永続化されるのを避けたい、という動機です。vim-jp slackで
は参加者の「忘れられる権利」を尊重しています。

以上の理由から消せる状態でデータ=差分をオンライン上にホストできる GitHub
Actions の Artifacts を利用しています。

## LICNESE

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="クリエイティブ・コモンズ・ライセンス" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />この 作品 は <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">クリエイティブ・コモンズ 表示 4.0 国際 ライセンス</a>の下に提供されています。
