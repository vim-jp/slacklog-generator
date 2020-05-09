.PHONY: build
build:
	go build

.PHONY: test
test:
	go test . ./internal/... ./subcmd/...

.PHONY: vet
vet:
	go vet . ./internal/... ./subcmd/...

.PHONY: lint
lint:
	golint . ./internal/... ./subcmd/...

_site: slacklog_pages
	./scripts/build.sh
	touch -c _site

slacklog_pages: slacklog_data $(wildcard scripts/**) $(wildcard templates/**)
	./scripts/generate_html.sh
	touch -c slacklog_pages

slacklog_data:
	curl -Ls https://github.com/vim-jp/slacklog/archive/log-data.tar.gz | tar xz --strip-components=1 --exclude=.github

.phony: clean
clean:
	rm -rf _site
	rm -rf emojis
	rm -rf files
	rm -rf slacklog_data
	rm -rf slacklog_pages
