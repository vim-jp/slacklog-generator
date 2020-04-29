# slacklog

## 開発に必要なもの

- Go
- ローカルで開発する場合 (Docker を使う場合は不要)
  - Ruby
  - Jekyll
  - (あれば)GNU Make

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
