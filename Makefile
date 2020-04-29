slacklog_pages: slacklog_data $(wildcard scripts/**) $(wildcard slacklog_template/**)
	./scripts/generate_html.sh
	touch -c slacklog_pages

slacklog_data:
	git fetch origin log-data
	git archive origin/log-data | tar x
