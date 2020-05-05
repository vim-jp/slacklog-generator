slacklog_pages: slacklog_data $(wildcard scripts/**) $(wildcard slacklog_template/**)
	./scripts/generate_html.sh
	touch -c slacklog_pages

slacklog_data:
	curl -Ls https://github.com/vim-jp/slacklog/archive/log-data.tar.gz | tar xz --strip-components=1 --exclude=.github

.phony: clean
clean:
	rm -rf emojis
	rm -rf files
	rm -rf slacklog_data
	rm -rf slacklog_pages
