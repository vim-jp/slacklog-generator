slacklog_pages: slacklog_data $(wildcard scripts/**) $(wildcard slacklog_template/**)
	./scripts/generate_html.sh
	touch -c slacklog_pages

slacklog_data:
	git fetch origin log-data
	git archive origin/log-data | tar x

.phony: clean
clean:
	rm -rf emojis
	rm -rf files
	rm -rf slacklog_data
	rm -rf slacklog_pages
