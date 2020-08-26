package slacklog

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/kyokomi/emoji"
	"github.com/slack-go/slack"
)

// HTMLGenerator : ログデータからHTMLを生成するための構造体。
type HTMLGenerator struct {
	// text/template形式のテンプレートが置いてあるディレクトリ
	templateDir string
	// files がおいてあるディレクトリ
	filesDir string
	// ログデータを取得するためのLogStore
	s *LogStore
	// markdown形式のテキストを変換するためのTextConverter
	c   *TextConverter
	cfg Config

	// baseURL is root path for public site, configured by `BASEURL` environment variable.
	baseURL string

	// ueMap is a set of unknown emojis.
	ueMap map[string]struct{}
	ueMu  sync.Mutex
}

// maxEmbeddedFileSize : 添付ファイルの埋め込みを行うファイルサイズ
// これ以下の場合、表示されるようになる
const maxEmbeddedFileSize = 102400

// NewHTMLGenerator : HTMLGeneratorを生成する。
func NewHTMLGenerator(templateDir string, filesDir string, s *LogStore) *HTMLGenerator {
	users := s.GetDisplayNameMap()
	emojis := s.GetEmojiMap()
	c := NewTextConverter(users, emojis)

	return &HTMLGenerator{
		templateDir: templateDir,
		filesDir:    filesDir,
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
	t, err := template.New(name).ParseFiles(tmplPath)
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
		} else if keys[i].year > keys[j].year {
			return false
		}
		return keys[i].month < keys[j].month
	})

	params := make(map[string]interface{})
	params["baseURL"] = g.baseURL
	params["channel"] = channel
	params["keys"] = keys

	tempPath := filepath.Join(g.templateDir, "channel_index.tmpl")
	name := filepath.Base(tempPath)
	t, err := template.New(name).ParseFiles(tempPath)
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

	t, err := template.New("").
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
				if msg.Username != "" {
					return g.c.escapeSpecialChars(msg.Username)
				}
				return g.c.escapeSpecialChars(g.s.GetDisplayNameByUserID(msg.User))
			},
			"userIconUrl": func(msg *Message) string {
				if msg.Icons != nil && msg.Icons.Image48 != "" {
					return msg.Icons.Image48
				}
				userID := msg.User
				if userID == "" && msg.BotID != "" {
					userID = msg.BotID
				}
				user, ok := g.s.GetUserByID(userID)
				if !ok {
					return "" // TODO show default icon
				}
				return user.Profile.Image48
			},
			"text":           g.generateMessageText,
			"reactions":      g.getReactions,
			"attachmentText": g.generateAttachmentText,
			"fileHTML":       g.generateFileHTML,
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
			"hostBySlack":      HostBySlack,
			"localPath":        LocalPath,
			"topLevelMimetype": TopLevelMimetype,
			"thumbImagePath":   ThumbImagePath,
			"thumbImageWidth":  ThumbImageWidth,
			"thumbImageHeight": ThumbImageHeight,
			"thumbVideoPath":   ThumbVideoPath,
			"stringsJoin":      strings.Join,
		}).
		ParseGlob(filepath.Join(g.templateDir, "channel_per_month", "*.tmpl"))
	if err != nil {
		return err
	}
	tmpl := t.Lookup("index.tmpl")
	if tmpl == nil {
		return errors.New("no index.tmpl in channel_per_month/ dir")
	}
	err = executeAndWrite(tmpl, params, filepath.Join(path, "index.html"))
	if err != nil {
		return err
	}
	return nil
}

func (g *HTMLGenerator) isVisibleMessage(msg Message) bool {
	return msg.SubType == "" || msg.SubType == "bot_message" || msg.SubType == "slackbot_response" || msg.SubType == "thread_broadcast"
}

func (g *HTMLGenerator) generateMessageText(msg Message) string {
	text := g.c.ToHTML(msg.Text)
	if msg.Edited != nil && g.cfg.EditedSuffix != "" {
		text += "<span class='slacklog-text-edited'>" + html.EscapeString(g.cfg.EditedSuffix) + "</span>"
	}
	return text
}

func (g *HTMLGenerator) generateAttachmentText(attachment slack.Attachment) string {
	return g.c.ToHTML(attachment.Text)
}

// generateFileHTML : 'text/plain' な添付ファイルをHTMLに埋め込む
// 存在しない場合、エラーを表示する
func (g *HTMLGenerator) generateFileHTML(file slack.File) string {
	if file.Size > maxEmbeddedFileSize {
		return `<span class="file-error">file size is too big to embed. please download from above link to see.</span>`
	}
	path := filepath.Join(g.filesDir, file.ID, LocalName(file, file.URLPrivate, ""))
	src, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf(`<span class="file-error">no files found: %s</span>`, err)
		}
		return fmt.Sprintf(`<span class="file-error">failed to read a file: %s</span>`, err)
	}
	ftype := file.Filetype
	if file.Filetype == "text" {
		ftype = "none"
	}
	return "<code class='language-" + ftype + "'>" + html.EscapeString(string(src)) + "</code>"
}

// ReactionInfo is information for a reaction.
type ReactionInfo struct {
	EmojiPath string
	Name      string
	Count     int
	Users     []string
	Default   bool
}

func (g *HTMLGenerator) getReactions(msg Message) []ReactionInfo {
	var info []ReactionInfo

	for _, reaction := range msg.Reactions {
		users := make([]string, 0, len(reaction.Users))
		for _, user := range reaction.Users {
			n := g.s.GetDisplayNameByUserID(user)
			if n == "" {
				continue
			}
			users = append(users, n)
		}

		// custom emoji case
		emojiExt, ok := g.s.et.NameToExt[reaction.Name]
		if ok {
			info = append(info, ReactionInfo{
				EmojiPath: url.PathEscape(reaction.Name + emojiExt),
				Name:      reaction.Name,
				Count:     reaction.Count,
				Users:     users,
				Default:   false,
			})
			continue
		}

		// fallback to unicode.
		emojiStr := ":" + reaction.Name + ":"
		unicodeEmojis := g.emojiToString(emojiStr)
		if unicodeEmojis == "" {
			// This may be a deleted emoji. Show `:emoji:` as is.
			unicodeEmojis = emojiStr
		}
		info = append(info, ReactionInfo{
			Name:    unicodeEmojis,
			Count:   reaction.Count,
			Users:   users,
			Default: true,
		})
	}

	return info
}

var rxEmoji = regexp.MustCompile(`:[^:]+:`)

func (g *HTMLGenerator) emojiToString(emojiSeq string) string {
	b := &strings.Builder{}
	for _, s := range rxEmoji.FindAllString(emojiSeq, -1) {
		ch, ok := emoji.CodeMap()[s]
		if !ok {
			g.ueMu.Lock()
			if g.ueMap == nil {
				g.ueMap = map[string]struct{}{}
			}
			if _, ok := g.ueMap[s]; !ok {
				g.ueMap[s] = struct{}{}
			}
			g.ueMu.Unlock()
			continue
		}
		b.WriteString(ch)
	}
	return b.String()
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
