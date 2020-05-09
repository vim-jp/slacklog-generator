_site: slacklog_data $(wildcard scripts/**) $(wildcard templates/**)
	./scripts/generate_html.sh
	./scripts/build.sh
	touch -c _site

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

slacklog_data:
	curl -Ls https://github.com/vim-jp/slacklog/archive/log-data.tar.gz | tar xz --strip-components=1 --exclude=.github

.PHONY: clean
clean:
	rm -rf _site

.PHONY: distclean
distclean: clean logdata-clean

##############################################################################
# manage logdata

.PHONY: logdata
logdata: _logdata

.PHONY: logdata-clean
logdata-clean:
	rm -rf logdata

.PHONY: logdata-distclean
logdata-distclean: logdata-clean
	rm -f tmp/log-data.tar.gz

.PHONY: logdata-restore
logdata-restore: logdata-clean logdata

.PHONY: logdata-update
logdata-update: logdata-distclean logdata

_logdata: tmp/log-data.tar.gz
	rm -rf $@
	mkdir -p $@
	tar xz --strip-components=1 --exclude=.github -f tmp/log-data.tar.gz -C $@

tmp/log-data.tar.gz:
	mkdir -p tmp
	curl -Lo $@ https://github.com/vim-jp/slacklog/archive/log-data.tar.gz
