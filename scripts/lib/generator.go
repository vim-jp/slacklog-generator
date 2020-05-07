package slacklog

import (
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"
)

// HTMLGenerator : ログデータからHTMLを生成するための構造体。
type HTMLGenerator struct {
	// text/template形式のテンプレートが置いてあるディレクトリ
	templateDir string
	// ログデータを取得するためのLogStore
	s *LogStore
	// markdown形式のテキストを変換するためのTextConverter
	c   *TextConverter
	cfg Config
	// 公開するサイトのURLのベース
	// 環境変数BASEURLで指定する
	baseURL string
}

// NewHTMLGenerator : HTMLGeneratorを生成する。
func NewHTMLGenerator(templateDir string, s *LogStore) *HTMLGenerator {
	users := s.GetDisplayNameMap()
	emojis := s.GetEmojiMap()
	c := NewTextConverter(users, emojis)

	return &HTMLGenerator{
		templateDir: templateDir,
		s:           s,
		c:           c,
		baseURL:     os.Getenv("BASEURL"),
	}
}

// Generate はoutDirにログデータの変換結果を生成する。
// 目標とする構造は以下となる:
//   - outDir/
//     - index.html // generateIndex()
//     - ${channel_id}/ // generateChannelDir()
//       - index.html // generateChannelIndex()
//       - ${YYYY}/
//         - ${MM}/
//           - index.html // generateMessageDir()
func (g *HTMLGenerator) Generate(outDir string) error {
	channels := g.s.GetChannels()

	createdChannels := []Channel{}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	for i := range channels {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			isCreated, err := g.generateChannelDir(
				filepath.Join(outDir, channels[i].ID),
				channels[i],
			)
			if err != nil {
				log.Printf("generateChannelDir(%s) failed: %s", channels[i].ID, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			if isCreated {
				mu.Lock()
				createdChannels = append(createdChannels, channels[i])
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs[0]
	}

	if err := g.generateIndex(filepath.Join(outDir, "index.html"), createdChannels); err != nil {
		return err
	}

	return nil
}

func (g *HTMLGenerator) generateIndex(path string, channels []Channel) error {
	params := make(map[string]interface{})
	SortChannel(channels)
	params["baseURL"] = g.baseURL
	params["channels"] = channels
	tmplPath := filepath.Join(g.templateDir, "index.tmpl")
	name := filepath.Base(tmplPath)
	t, err := template.New(name).
		Funcs(map[string]interface{}{
			"jekyll_through": jekyllThrough,
			"J":              jekyllThrough, // shorthand for "jekyll_through"
		}).
		ParseFiles(tmplPath)
	if err != nil {
		return err
	}
	if err := executeAndWrite(t, params, path); err != nil {
		return err
	}
	return nil
}

func (g *HTMLGenerator) generateChannelDir(path string, channel Channel) (bool, error) {
	msgsMap, err := g.s.GetMessagesPerMonth(channel.ID)
	if err != nil {
		return false, err
	}
	if len(msgsMap) == 0 {
		return false, nil
	}

	if err := os.MkdirAll(path, 0777); err != nil {
		return false, fmt.Errorf("could not create %s directory: %w", path, err)
	}

	if err := g.generateChannelIndex(
		channel,
		msgsMap.Keys(),
		filepath.Join(path, "index.html"),
	); err != nil {
		return true, err
	}

	for key, mm := range msgsMap {
		if err := g.generateMessageDir(
			channel,
			key,
			mm,
			filepath.Join(path, key.Year(), key.Month()),
		); err != nil {
			return true, err
		}
	}
	return true, nil
}

func (g *HTMLGenerator) generateChannelIndex(channel Channel, keys []MessageMonthKey, path string) error {

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].year < keys[j].year {
			return true
		}
		return keys[i].month < keys[j].month
	})

	params := make(map[string]interface{})
	params["baseURL"] = g.baseURL
	params["channel"] = channel
	params["keys"] = keys

	tempPath := filepath.Join(g.templateDir, "channel_index.tmpl")
	name := filepath.Base(tempPath)
	t, err := template.New(name).
		Funcs(map[string]interface{}{
			"jekyll_through": jekyllThrough,
			"J":              jekyllThrough, // shorthand for "jekyll_through"
		}).
		ParseFiles(tempPath)
	if err != nil {
		return err
	}
	if err := executeAndWrite(t, params, path); err != nil {
		return err
	}
	return nil
}

func (g *HTMLGenerator) generateMessageDir(channel Channel, key MessageMonthKey, msgs Messages, path string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return fmt.Errorf("could not create %s directory: %w", path, err)
	}

	params := make(map[string]interface{})
	params["baseURL"] = g.baseURL
	params["channel"] = channel
	params["monthKey"] = key
	params["msgs"] = msgs

	// TODO check below subtypes work correctly
	// TODO support more subtypes
	tmplPath := filepath.Join(g.templateDir, "channel_per_month_index.tmpl")
	name := filepath.Base(tmplPath)
	t, err := template.New(name).
		Funcs(map[string]interface{}{
			"visible": g.isVisibleMessage,
			"datetime": func(ts string) string {
				return TsToDateTime(ts).Format("2日 15:04:05")
			},
			"threadMessageTime": func(msgTs, threadTs string) string {
				return LevelOfDetailTime(TsToDateTime(msgTs), TsToDateTime(threadTs))
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
				if t, ok := g.s.GetThread(channel.ID, ts); ok {
					return LevelOfDetailTime(t.LastReplyTime(), TsToDateTime(ts))
				}
				return ""
			},
			"threads": func(ts string) Messages {
				if t, ok := g.s.GetThread(channel.ID, ts); ok {
					return t.Replies()
				}
				return nil
			},
			"threadNum": func(ts string) int {
				if t, ok := g.s.GetThread(channel.ID, ts); ok {
					return t.ReplyCount()
				}
				return 0
			},
			"threadRootText": func(ts string) string {
				thread, ok := g.s.GetThread(channel.ID, ts)
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
			"hasPrevMonth": func(key MessageMonthKey) bool {
				return g.s.HasPrevMonth(channel.ID, key)
			},
			"hasNextMonth": func(key MessageMonthKey) bool {
				return g.s.HasNextMonth(channel.ID, key)
			},
			"jekyll_through": jekyllThrough,
			"J":              jekyllThrough, // shorthand for "jekyll_through"
		}).ParseFiles(tmplPath)
	if err != nil {
		return err
	}
	err = executeAndWrite(t, params, filepath.Join(path, "index.html"))
	if err != nil {
		return err
	}
	return nil
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

// executeAndWrite executes a template and writes contents to a file.
func executeAndWrite(tmpl *template.Template, data interface{}, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}
	return nil
}

func jekyllThrough(s string) string {
	return "{{ " + s + " }}"
}
