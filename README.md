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
git fetch origin log-data
git archive origin/log-data | tar x
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

### geneate-html 差分コマンドの出力の差分の確認方法

以下のコマンドで自分が変更した結果として変化した generate-html の出力内容の差分
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

## LICNESE

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="クリエイティブ・コモンズ・ライセンス" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />この 作品 は <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">クリエイティブ・コモンズ 表示 4.0 国際 ライセンス</a>の下に提供されています。
