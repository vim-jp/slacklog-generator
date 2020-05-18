PACKAGES=./...

.PHONY: generate
generate: _site

_site: _logdata $(wildcard scripts/**) $(wildcard templates/**)
	./scripts/generate_html.sh
	./scripts/build.sh
	touch -c _site

.PHONY: clean
clean: go-clean
	rm -rf _site

.PHONY: distclean
distclean: clean logdata-distclean

##############################################################################
# Go

.PHONY: build
build:
	go build

.PHONY: test
test:
	go test ${PACKAGES}

# cover - テストを実行してカバレッジを計測し結果を tmp/cover.html に出力する
.PHONY: cover
cover:
	go test -coverprofile tmp/cover.out ${PACKAGES}
	go tool cover -html tmp/cover.out -o tmp/cover.html

.PHONY: vet
vet:
	go vet ${PACKAGES}

.PHONY: lint
lint:
	golint ${PACKAGES}

.PHONY: go-clean
go-clean:
	go clean

##############################################################################
# manage logdata

.PHONY: logdata
logdata: _logdata

.PHONY: logdata-clean
logdata-clean:
	rm -rf _logdata

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
