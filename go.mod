module github.com/vim-jp/slacklog-generator

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kyokomi/emoji v2.2.2+incompatible
	github.com/pkg/errors v0.9.1 // indirect
	github.com/slack-go/slack v0.6.4
	github.com/urfave/cli/v2 v2.2.0
)

replace github.com/slack-go/slack => ../slacklog-slack
