package slacklog

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

type HTMLGenerator struct {
	templateDir string
	s           *LogStore
	c           *TextConverter
	cfg         Config
}

func NewHTMLGenerator(templateDir string, s *LogStore) *HTMLGenerator {
	users := s.GetUserNameMap()
	emojis := s.GetEmojiMap()
	c := NewTextConverter(users, emojis)

	return &HTMLGenerator{
		templateDir: templateDir,
		s:           s,
		c:           c,
	}
}

func (g *HTMLGenerator) Generate(outDir string) error {
	channels := g.s.GetChannels()

	for i := range channels {
		if err := g.generateChannelDir(
			filepath.Join(outDir, channels[i].ID),
			channels[i],
		); err != nil {
			return err
		}
	}
	return nil
}

func (g *HTMLGenerator) generateChannelDir(path string, channel Channel) error {
	msgs, err := g.s.GetMessagesPerMonth(channel.ID)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	if err := mkdir(path); err != nil {
		return fmt.Errorf("could not create %s directory: %s", path, err)
	}

	if err := g.generateChannelIndex(
		channel,
		msgs,
		filepath.Join(path, "index.html"),
	); err != nil {
		return err
	}

	for i := range msgs {
		if err := g.generateMessageDir(
			channel,
			msgs[i],
			filepath.Join(path, msgs[i].Year(), msgs[i].Month()),
		); err != nil {
			return err
		}
	}
	return nil
}

func (g *HTMLGenerator) generateChannelIndex(channel Channel, msgs []MessagesPerMonth, path string) error {
	params := make(map[string]interface{})
	params["channel"] = channel
	params["msgMap"] = msgs
	var out bytes.Buffer

	tempPath := filepath.Join(g.templateDir, "channel_index.tmpl")
	name := filepath.Base(tempPath)
	t, err := template.New(name).Delims("<<", ">>").ParseFiles(tempPath)
	if err != nil {
		return err
	}
	if err := t.Execute(&out, params); err != nil {
		return err
	}
	return ioutil.WriteFile(path, out.Bytes(), 0666)
}

func (g *HTMLGenerator) generateMessageDir(channel Channel, msgs MessagesPerMonth, path string) error {
	if err := mkdir(path); err != nil {
		return fmt.Errorf("could not create %s directory: %s", path, err)
	}

	// TODO: impl
	var threadMap map[string]Thread

	params := make(map[string]interface{})
	params["channel"] = channel
	params["msgPerMonth"] = msgs
	params["threadMap"] = threadMap
	var out bytes.Buffer

	// TODO check below subtypes work correctly
	// TODO support more subtypes
	tmplPath := filepath.Join(g.templateDir, "channel_per_month_index.tmpl")
	name := filepath.Base(tmplPath)
	t, err := template.New(name).
		Delims("<<", ">>").
		Funcs(map[string]interface{}{
			"visible": g.isVisibleMessage,
			"datetime": func(ts string) string {
				return TsToDateTime(ts).Format("2日 15:04:05")
			},
			"slackPermalink": func(ts string) string {
				return strings.Replace(ts, ".", "", 1)
			},
			"username": func(msg *Message) string {
				if msg.Subtype == "bot_message" || msg.Subtype == "slackbot_response" {
					return g.c.escapeSpecialChars(msg.Username)
				}
				return g.c.escapeSpecialChars(g.s.GetDisplayNameByUserID(msg.User))
			},
			"userIconUrl": func(msg *Message) string {
				switch msg.Subtype {
				case "", "thread_broadcast":
					user, ok := g.s.GetUserByID(msg.User)
					if !ok {
						return "" // TODO show default icon
					}
					return user.Profile.Image48
				case "bot_message", "slackbot_response":
					if msg.Icons != nil && msg.Icons.Image48 != "" {
						return msg.Icons.Image48
					}
				}
				return ""
			},
			"text":           g.generateMessageText,
			"attachmentText": g.generateAttachmentText,
			"threadMtime": func(ts string) string {
				if t, ok := g.s.GetThread(ts); ok {
					return t.LastReplyTime().Format("2日 15:04:05")
				}
				return ""
			},
			"threads": func(ts string) []Message {
				if t, ok := g.s.GetThread(ts); ok {
					return t.Replies()
				}
				return nil
			},
			"threadNum": func(ts string) int {
				if t, ok := g.s.GetThread(ts); ok {
					return t.ReplyNum()
				}
				return 0
			},
			"threadRootText": func(ts string) string {
				thread, ok := g.s.GetThread(ts)
				if !ok {
					return ""
				}
				runes := []rune(thread.RootText())
				text := string(runes)
				if len(runes) > 20 {
					text = string(runes[:20]) + " ..."
				}
				return g.c.escape(text)
			},
			"hasPrevMonth": func(msgs MessagesPerMonth) bool {
				return g.s.HasPrevMonth(msgs)
			},
			"hasNextMonth": func(msgs MessagesPerMonth) bool {
				return g.s.HasNextMonth(msgs)
			},
		}).ParseFiles(tmplPath)
	if err != nil {
		return err
	}
	if err := t.Execute(&out, params); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(path, "index.html"), out.Bytes(), 0666)
}

func (g *HTMLGenerator) isVisibleMessage(msg Message) bool {
	return msg.Subtype == "" || msg.Subtype == "bot_message" || msg.Subtype == "slackbot_response" || msg.Subtype == "thread_broadcast"
}

func (g *HTMLGenerator) generateMessageText(msg Message) string {
	text := g.c.ToHTML(msg.Text)
	if msg.Edited != nil && g.cfg.EditedSuffix != "" {
		text += "<span class='slacklog-text-edited'>" + html.EscapeString(g.cfg.EditedSuffix) + "</span>"
	}
	return text
}

func (g *HTMLGenerator) generateAttachmentText(attachment MessageAttachment) string {
	return g.c.ToHTML(attachment.Text)
}
