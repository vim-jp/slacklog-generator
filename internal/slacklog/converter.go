package slacklog

import (
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/kyokomi/emoji"
)

// TextConverter : markdown形式のテキストをHTMLに変換するための構造体。
type TextConverter struct {
	// key: emoji name
	// value: emoji URL
	emojis map[string]string
	// key: user ID
	// value: display name
	users map[string]string
	re    regexps
	// baseURL is root path for public site, configured by `BASEURL` environment variable.
	baseURL string
}

// NewTextConverter : TextConverter を生成する
func NewTextConverter(users, emojis map[string]string) *TextConverter {
	re := regexps{}
	// TODO tokenize/parse message.Text
	re.linkWithTitle = regexp.MustCompile(`&lt;(https?://[^>]+?\|(.+?))&gt;`)
	re.link = regexp.MustCompile(`&lt;(https?://[^>]+?)&gt;`)
	// go regexp does not support back reference
	re.code = regexp.MustCompile("`{3}|｀{3}")
	re.codeShort = regexp.MustCompile("[`｀]([^`]+?)[`｀]")
	re.del = regexp.MustCompile(`~([^~]+?)~`)
	re.mention = regexp.MustCompile(`&lt;@(\w+?)&gt;`)
	re.channel = regexp.MustCompile(`&lt;#([^|]+?)\|([^&]+?)&gt;`)
	re.emoji = regexp.MustCompile(`:[^\s!"#$%&()=^/?\\\[\]<>,.;@{}~:]+:`)
	re.newLine = regexp.MustCompile(`\n`)

	return &TextConverter{
		emojis:  emojis,
		users:   users,
		re:      re,
		baseURL: os.Getenv("BASEURL"),
	}
}

type regexps struct {
	linkWithTitle, link,
	code, codeShort,
	del,
	mention, channel, emoji,
	newLine *regexp.Regexp
}

func (c *TextConverter) escapeSpecialChars(text string) string {
	text = html.EscapeString(html.UnescapeString(text))
	text = strings.Replace(text, "{{", "&#123;&#123;", -1)
	return strings.Replace(text, "{%", "&#123;&#37;", -1)
}

func (c *TextConverter) escape(text string) string {
	text = html.EscapeString(html.UnescapeString(text))
	text = c.re.newLine.ReplaceAllString(text, " ")
	return text
}

func (c *TextConverter) bindEmoji(emojiExp string) string {
	name := emojiExp[1 : len(emojiExp)-1]
	extension, ok := c.emojis[name]
	if !ok {
		char, ok := emoji.CodeMap()[emojiExp]
		if ok {
			return char
		}
		return emojiExp
	}
	for 7 <= len(extension) && extension[:6] == "alias:" {
		name = extension[6:]
		extension, ok = c.emojis[name]
		if !ok {
			return emojiExp
		}
	}
	src := c.baseURL + "/emojis/" + url.PathEscape(name) + extension
	return "<img class='slacklog-emoji' title='" + emojiExp + "' alt='" + emojiExp + "' src='" + src + "'>"
}

func (c *TextConverter) bindUser(userExp string) string {
	m := c.re.mention.FindStringSubmatch(userExp)
	if name := c.users[m[1]]; name != "" {
		return "@" + name
	}
	return userExp
}

func (c *TextConverter) bindChannel(channelExp string) string {
	matchResult := c.re.channel.FindStringSubmatch(channelExp)
	channelID := matchResult[1]
	channelName := matchResult[2]
	return "<a href='" + c.baseURL + "/" + channelID + "/'>#" + channelName + "</a>"
}

// ToHTML : markdown形式のtextをHTMLに変換する
func (c *TextConverter) ToHTML(text string) string {
	text = c.escapeSpecialChars(text)
	text = c.re.newLine.ReplaceAllString(text, "<br>")
	chunks := c.re.code.Split(text, -1)
	for i, s := range chunks {
		if i%2 == 0 {
			s = c.re.linkWithTitle.ReplaceAllString(s, "<a href='${1}'>${2}</a>")
			s = c.re.link.ReplaceAllString(s, "<a href='${1}'>${1}</a>")
			s = c.re.codeShort.ReplaceAllString(s, "<code>${1}</code>")
			s = c.re.del.ReplaceAllString(s, "<del>${1}</del>")
			s = c.re.emoji.ReplaceAllStringFunc(s, c.bindEmoji)
			s = c.re.mention.ReplaceAllStringFunc(s, c.bindUser)
			s = c.re.channel.ReplaceAllStringFunc(s, c.bindChannel)
		} else {
			s = "<pre>" + s + "</pre>"
		}
		chunks[i] = s
	}
	return strings.Join(chunks, "")
}
