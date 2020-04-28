package slacklog

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/kyokomi/emoji"
)

type TextConverter struct {
	emojis map[string]string
	users  map[string]string
	re     regexps
}

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
		emojis: emojis,
		users:  users,
		re:     re,
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
	for 6 <= len(extension) && extension[:6] == "alias:" {
		name = extension[6:]
		extension, ok = c.emojis[name]
		if !ok {
			return emojiExp
		}
	}
	src := "{{ site.baseurl }}/emojis/" + url.PathEscape(name) + extension
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
	channelId := matchResult[1]
	channelName := matchResult[2]
	return "<a href='{{ site.baseurl }}/" + channelId + "/'>#" + channelName + "</a>"
}

func (c *TextConverter) ToHTML(text string) string {
	text = c.escapeSpecialChars(text)
	text = c.re.newLine.ReplaceAllString(text, "<br>")
	chunks := c.re.code.Split(text, -1)
	for i := range chunks {
		if i%2 == 0 {
			chunks[i] = c.re.linkWithTitle.ReplaceAllString(chunks[i], "<a href='${1}'>${2}</a>")
			chunks[i] = c.re.link.ReplaceAllString(chunks[i], "<a href='${1}'>${1}</a>")
			chunks[i] = c.re.codeShort.ReplaceAllString(chunks[i], "<code>${1}</code>")
			chunks[i] = c.re.del.ReplaceAllString(chunks[i], "<del>${1}</del>")
			chunks[i] = c.re.emoji.ReplaceAllStringFunc(chunks[i], c.bindEmoji)
			chunks[i] = c.re.mention.ReplaceAllStringFunc(chunks[i], c.bindUser)
			chunks[i] = c.re.channel.ReplaceAllStringFunc(chunks[i], c.bindChannel)
		} else {
			chunks[i] = "<pre>" + chunks[i] + "</pre>"
		}
	}
	return strings.Join(chunks, "")
}
