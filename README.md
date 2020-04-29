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

#### 添付ファイルと絵文字のダウンロード

```console
export SLACK_TOKEN=xxxx
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

## LICNESE

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="クリエイティブ・コモンズ・ライセンス" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />この 作品 は <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">クリエイティブ・コモンズ 表示 4.0 国際 ライセンス</a>の下に提供されています。
