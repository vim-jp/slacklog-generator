slacklog_pages: slacklog_data scripts/config.json scripts/main.go scripts/lib/generate_html.go $(wildcard slacklog_template/**)
	./scripts/generate_html.sh
	touch --no-create slacklog_pages

slacklog_data:
	git fetch origin log-data
	git archive origin/log-data | tar x
