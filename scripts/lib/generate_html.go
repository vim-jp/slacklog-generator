package slacklog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kyokomi/emoji"
)

func doGenerateHTML() error {
	if len(os.Args) < 6 {
		fmt.Println("Usage: go run scripts/main.go generate_html {config.json} {templatedir} {indir} {outdir}")
		return nil
	}

	configJsonPath := filepath.Clean(os.Args[2])
	templateDir := filepath.Clean(os.Args[3])
	inDir := filepath.Clean(os.Args[4])
	outDir := filepath.Clean(os.Args[5])

	cfg, err := readConfig(configJsonPath)
	if err != nil {
		return fmt.Errorf("could not read config: %s", err)
	}
	_, userMap, err := readUsers(filepath.Join(inDir, "users.json"))
	if err != nil {
		return fmt.Errorf("could not read users.json: %s", err)
	}
	channels, _, err := readChannels(filepath.Join(inDir, "channels.json"), cfg.Channels)
	if err != nil {
		return fmt.Errorf("could not read channels.json: %s", err)
	}

	emojis := readEmojiJson(cfg, configJsonPath)

	if err := mkdir(outDir); err != nil {
		return fmt.Errorf("could not create out directory: %s", err)
	}

	emptyChannel := make(map[string]bool, len(channels))
	for i := range channels {
		if err := mkdir(filepath.Join(outDir, channels[i].Id)); err != nil {
			return fmt.Errorf("could not create %s/%s directory: %s", outDir, channels[i].Id, err)
		}
		msgKv, threadMap, err := getMsgPerMonth(inDir, channels[i].Id)
		if msgKv.isEmpty() {
			emptyChannel[channels[i].Id] = true
			continue
		}
		if err != nil {
			return err
		}
		// Generate {outdir}/{channel}/index.html (links to {channel}/{year}/{month})
		content, err := genChannelIndex(inDir, filepath.Join(templateDir, "channel_index.tmpl"), &channels[i], msgKv, cfg)
		if err != nil {
			return fmt.Errorf("could not generate %s/%s: %s", outDir, channels[i].Id, err)
		}
		err = ioutil.WriteFile(filepath.Join(outDir, channels[i].Id, "index.html"), content, 0666)
		if err != nil {
			return fmt.Errorf("could not create %s/%s/index.html: %s", outDir, channels[i].Id, err)
		}
		// Generate {outdir}/{channel}/{year}/{month}/index.html
		for _, msgPerMonth := range msgKv.Enumerate() {
			if err := mkdir(filepath.Join(outDir, channels[i].Id, msgPerMonth.Year(), msgPerMonth.Month())); err != nil {
				return fmt.Errorf("could not create %s/%s/%s/%s directory: %s", outDir, channels[i].Id, msgPerMonth.Year(), msgPerMonth.Month(), err)
			}
			content, err := genChannelPerMonthIndex(inDir, filepath.Join(templateDir, "channel_per_month_index.tmpl"), &channels[i], msgPerMonth, userMap, threadMap, emojis, cfg)
			if err != nil {
				return fmt.Errorf("could not generate %s/%s/%s/%s/index.html: %s", outDir, channels[i].Id, msgPerMonth.Year(), msgPerMonth.Month(), err)
			}
			err = ioutil.WriteFile(filepath.Join(outDir, channels[i].Id, msgPerMonth.Year(), msgPerMonth.Month(), "index.html"), content, 0666)
			if err != nil {
				return fmt.Errorf("could not create %s/%s/index.html: %s", outDir, channels[i].Id, err)
			}
		}
	}

	// Remove empty channels
	newChannels := make([]channel, 0, len(channels))
	for i := range channels {
		if !emptyChannel[channels[i].Id] {
			newChannels = append(newChannels, channels[i])
		}
	}
	channels = newChannels

	// Generate {outdir}/index.html (links to {channel})
	content, err := genIndex(channels, filepath.Join(templateDir, "index.tmpl"), cfg)
	err = ioutil.WriteFile(filepath.Join(outDir, "index.html"), content, 0666)
	if err != nil {
		return fmt.Errorf("could not create %s/index.html: %s", outDir, err)
	}

	return nil
}

func mkdir(path string) error {
	os.MkdirAll(path, 0777)
	if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
		return err
	}
	return nil
}

func visibleMsg(msg *message) bool {
	return msg.Subtype == "" || msg.Subtype == "bot_message" || msg.Subtype == "slackbot_response" || msg.Subtype == "thread_broadcast"
}

func genIndex(channels []channel, tmplFile string, cfg *config) ([]byte, error) {
	params := make(map[string]interface{})
	params["channels"] = channels
	var out bytes.Buffer
	name := filepath.Base(tmplFile)
	t, err := template.New(name).Delims("<<", ">>").ParseFiles(tmplFile)
	if err != nil {
		return nil, err
	}
	err = t.Execute(&out, params)
	return out.Bytes(), err
}

func genChannelIndex(inDir, tmplFile string, channel *channel, msgKv *msgKeyValue, cfg *config) ([]byte, error) {
	params := make(map[string]interface{})
	params["channel"] = channel
	params["msgKv"] = msgKv
	var out bytes.Buffer
	name := filepath.Base(tmplFile)
	t, err := template.New(name).Delims("<<", ">>").ParseFiles(tmplFile)
	if err != nil {
		return nil, err
	}
	err = t.Execute(&out, params)
	return out.Bytes(), err
}

func genChannelPerMonthIndex(inDir, tmplFile string, channel *channel, msgPerMonth *msgEnum, userMap map[string]*user, threadMap msgThreadMap, emojis map[string]string, cfg *config) ([]byte, error) {
	params := make(map[string]interface{})
	params["channel"] = channel
	params["msgPerMonth"] = msgPerMonth
	params["threadMap"] = threadMap
	var out bytes.Buffer

	// TODO tokenize/parse message.Text
	var reLinkWithTitle = regexp.MustCompile(`&lt;(https?://[^>]+?\|(.+?))&gt;`)
	var reLink = regexp.MustCompile(`&lt;(https?://[^>]+?)&gt;`)
	// go regexp does not support back reference
	var reCode = regexp.MustCompile("`{3}|｀{3}")
	var reCodeShort = regexp.MustCompile("[`｀]([^`]+?)[`｀]")
	var reDel = regexp.MustCompile(`~([^~]+?)~`)
	var reMention = regexp.MustCompile(`&lt;@(\w+?)&gt;`)
	var reChannel = regexp.MustCompile(`&lt;#([^|]+?)\|([^&]+?)&gt;`)
	var reEmoji = regexp.MustCompile(`:[^\s!"#$%&()=^/?\\\[\]<>,.;@{}~:]+:`)
	var reNewline = regexp.MustCompile(`\n`)
	var escapeSpecialChars = func(text string) string {
		text = html.EscapeString(html.UnescapeString(text))
		text = strings.Replace(text, "{{", "&#123;&#123;", -1)
		return strings.Replace(text, "{%", "&#123;&#37;", -1)
	}
	var text2Html = func(text string) string {
		text = escapeSpecialChars(text)
		text = reNewline.ReplaceAllString(text, "<br>")
		chunks := reCode.Split(text, -1)
		for i := range chunks {
			if i%2 == 0 {
				chunks[i] = reLinkWithTitle.ReplaceAllString(chunks[i], "<a href='${1}'>${2}</a>")
				chunks[i] = reLink.ReplaceAllString(chunks[i], "<a href='${1}'>${1}</a>")
				chunks[i] = reCodeShort.ReplaceAllString(chunks[i], "<code>${1}</code>")
				chunks[i] = reDel.ReplaceAllString(chunks[i], "<del>${1}</del>")
				chunks[i] = reEmoji.ReplaceAllStringFunc(chunks[i], func(whole string) string {
					name := whole[1 : len(whole)-1]
					extension, ok := emojis[name]
					if !ok {
						char, ok := emoji.CodeMap()[whole]
						if ok {
							return char
						}
						return whole
					}
					for 6 <= len(extension) && extension[:6] == "alias:" {
						name = extension[6:]
						extension, ok = emojis[name]
						if !ok {
							return whole
						}
					}
					src := "{{ site.baseurl }}/emojis/" + url.PathEscape(name) + extension
					return "<img class='slacklog-emoji' title='" + whole + "' alt='" + whole + "' src='" + src + "'>"
				})
				chunks[i] = reMention.ReplaceAllStringFunc(chunks[i], func(whole string) string {
					m := reMention.FindStringSubmatch(whole)
					if name := getDisplayNameByUserId(m[1], userMap); name != "" {
						return "@" + name
					}
					return whole
				})
				chunks[i] = reChannel.ReplaceAllStringFunc(chunks[i], func(whole string) string {
					matchResult := reChannel.FindStringSubmatch(whole)
					channelId := matchResult[1]
					channelName := matchResult[2]
					return "<a href='{{ site.baseurl }}/" + channelId + "/'>#" + channelName + "</a>"
				})
			} else {
				chunks[i] = "<pre>" + chunks[i] + "</pre>"
			}
		}
		return strings.Join(chunks, "")
	}
	var escapeText = func(text string) string {
		text = html.EscapeString(html.UnescapeString(text))
		text = reNewline.ReplaceAllString(text, " ")
		return text
	}
	var funcText = func(msg *message) string {
		text := text2Html(msg.Text)
		if msg.Edited != nil && cfg.EditedSuffix != "" {
			text += "<span class='slacklog-text-edited'>" + html.EscapeString(cfg.EditedSuffix) + "</span>"
		}
		return text
	}
	var funcAttachmentText = func(attachment *messageAttachment) string {
		return text2Html(attachment.Text)
	}
	var ts2threadMtime = func(ts string) time.Time {
		lastMsg := threadMap[ts][len(threadMap[ts])-1]
		return ts2datetime(lastMsg.Ts)
	}

	// TODO check below subtypes work correctly
	// TODO support more subtypes
	name := filepath.Base(tmplFile)
	t, err := template.New(name).
		Delims("<<", ">>").
		Funcs(map[string]interface{}{
			"visible": visibleMsg,
			"datetime": func(ts string) string {
				return ts2datetime(ts).Format("2日 15:04:05")
			},
			"username": func(msg *message) string {
				if msg.Subtype == "bot_message" || msg.Subtype == "slackbot_response" {
					return escapeSpecialChars(msg.Username)
				}
				return escapeSpecialChars(getDisplayNameByUserId(msg.User, userMap))
			},
			"userIconUrl": func(msg *message) string {
				switch msg.Subtype {
				case "", "thread_broadcast":
					user, ok := userMap[msg.User]
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
			"text":           funcText,
			"attachmentText": funcAttachmentText,
			"threadMtime": func(ts string) string {
				return ts2threadMtime(ts).Format("2日 15:04:05")
			},
			"threads": func(ts string) []*message {
				if threads, ok := threadMap[ts]; ok {
					return threads[1:]
				}
				return nil
			},
			"threadNum": func(ts string) int {
				return len(threadMap[ts]) - 1
			},
			"threadRootText": func(ts string) string {
				threads, ok := threadMap[ts]
				if !ok {
					return ""
				}
				runes := []rune(threads[0].Text)
				text := string(runes)
				if len(runes) > 20 {
					text = string(runes[:20]) + " ..."
				}
				return escapeText(text)
			},
		}).
		ParseFiles(tmplFile)
	if err != nil {
		return nil, err
	}
	err = t.Execute(&out, params)
	return out.Bytes(), err
}

func ts2datetime(ts string) time.Time {
	t := strings.Split(ts, ".")
	if len(t) != 2 {
		return time.Time{}
	}
	sec, err := strconv.ParseInt(t[0], 10, 64)
	if err != nil {
		return time.Time{}
	}
	nsec, err := strconv.ParseInt(t[1], 10, 64)
	if err != nil {
		return time.Time{}
	}
	japan, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}
	}
	return time.Unix(sec, nsec).In(japan)
}

func getDisplayNameByUserId(userId string, userMap map[string]*user) string {
	if user, ok := userMap[userId]; ok {
		if user.Profile.RealName != "" {
			return user.Profile.RealName
		}
		if user.Profile.DisplayName != "" {
			return user.Profile.DisplayName
		}
	}
	return ""
}

type msgKeyValue struct {
	msgMap map[string][]message
}

type msgThreadMap map[string][]*message

type msgEnum struct {
	year     int
	month    int
	Messages []message
	kv       *msgKeyValue
}

func (kv *msgKeyValue) remove(me *msgEnum) {
	delete(kv.msgMap, kv.packKey(me.year, me.month))
}

func newMsgKeyValue() *msgKeyValue {
	return &msgKeyValue{msgMap: make(map[string][]message)}
}

func (kv *msgKeyValue) Enumerate() []*msgEnum {
	results := make([]*msgEnum, 0, len(kv.msgMap))
	for key := range kv.msgMap {
		year, month := kv.unpackKey(key)
		results = append(results, &msgEnum{
			year:     year,
			month:    month,
			Messages: kv.msgMap[key],
			kv:       kv,
		})
	}
	sort.SliceStable(results, func(i, j int) bool {
		n1 := results[i].year*100 + results[i].month
		n2 := results[j].year*100 + results[j].month
		return n1 < n2
	})
	return results
}

func (kv *msgKeyValue) packKey(year, month int) string {
	return fmt.Sprintf("%4d%02d", year, month)
}

func (kv *msgKeyValue) unpackKey(key string) (year, month int) {
	yyyy, mm := key[:4], key[4:6]
	year64, err := strconv.ParseInt(yyyy, 10, 32)
	if err != nil {
		panic(err)
	}
	month64, err := strconv.ParseInt(mm, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(year64), int(month64)
}

func (kv *msgKeyValue) hasEntry(year, month int) bool {
	_, ok := kv.msgMap[kv.packKey(year, month)]
	return ok
}

func (kv *msgKeyValue) isEmpty() bool {
	return len(kv.msgMap) == 0
}

func (kv *msgKeyValue) getMessagesByMonth(year, month int) []message {
	return kv.msgMap[kv.packKey(year, month)]
}

func (kv *msgKeyValue) createEmptyEntry(year, month int) {
	kv.msgMap[kv.packKey(year, month)] = []message{}
}

func (kv *msgKeyValue) appendMessagesByMonth(year, month int, msgs []message) {
	key := kv.packKey(year, month)
	kv.msgMap[key] = append(kv.msgMap[key], msgs...)
}

func (me *msgEnum) Year() string {
	return fmt.Sprintf("%4d", me.year)
}

func (me *msgEnum) Month() string {
	return fmt.Sprintf("%02d", me.month)
}

func (me *msgEnum) prevYearMonth() (year, month int) {
	if me.month <= 1 {
		return me.year - 1, 12
	}
	return me.year, me.month - 1
}

func (me *msgEnum) nextYearMonth() (year, month int) {
	if me.month >= 12 {
		return me.year + 1, 1
	}
	return me.year, me.month + 1
}

func (me *msgEnum) PrevYear() string {
	year, _ := me.prevYearMonth()
	return fmt.Sprintf("%4d", year)
}

func (me *msgEnum) PrevMonth() string {
	_, month := me.prevYearMonth()
	return fmt.Sprintf("%02d", month)
}

func (me *msgEnum) NextYear() string {
	year, _ := me.nextYearMonth()
	return fmt.Sprintf("%4d", year)
}

func (me *msgEnum) NextMonth() string {
	_, month := me.nextYearMonth()
	return fmt.Sprintf("%02d", month)
}

func (me *msgEnum) HasPrev() bool {
	year, month := me.prevYearMonth()
	msgs := me.kv.getMessagesByMonth(year, month)
	return len(msgs) != 0
}

func (me *msgEnum) HasNext() bool {
	year, month := me.nextYearMonth()
	msgs := me.kv.getMessagesByMonth(year, month)
	return len(msgs) != 0
}

// "{year}-{month}-{day}.json"
var reMsgFilename = regexp.MustCompile(`^(\d{4})-(\d{2})-\d{2}\.json$`)

func getMsgPerMonth(inDir string, channelName string) (*msgKeyValue, msgThreadMap, error) {
	dir, err := os.Open(filepath.Join(inDir, channelName))
	if err != nil {
		return nil, nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(names)
	msgKv := newMsgKeyValue()
	threadMap := make(msgThreadMap)
	for i := range names {
		m := reMsgFilename.FindStringSubmatch(names[i])
		if len(m) == 0 {
			fmt.Fprintf(os.Stderr, "[warning] skipping %s/%s/%s ...", inDir, channelName, names[i])
			continue
		}
		year64, err := strconv.ParseInt(m[1], 10, 32)
		if err != nil {
			return nil, nil, err
		}
		month64, err := strconv.ParseInt(m[2], 10, 32)
		if err != nil {
			return nil, nil, err
		}
		year, month := int(year64), int(month64)
		if !msgKv.hasEntry(year, month) {
			msgKv.createEmptyEntry(year, month)
		}
		msgs, err := readMessages(filepath.Join(inDir, channelName, names[i]), threadMap)
		if err != nil {
			return nil, nil, err
		}
		msgKv.appendMessagesByMonth(year, month, msgs)
	}
	for _, msgPerMonth := range msgKv.Enumerate() {
		m := msgPerMonth.Messages
		if len(m) == 0 {
			msgKv.remove(msgPerMonth)
			continue
		}
		sort.SliceStable(m, func(i, j int) bool {
			// must be the same digits, so no need to convert the timestamp to a number
			return m[i].Ts < m[j].Ts
		})
		var lastUser string
		for i := range m {
			if lastUser == m[i].User {
				m[i].Trail = true
			} else {
				lastUser = m[i].User
			}
		}
	}
	return msgKv, threadMap, nil
}

type message struct {
	ClientMsgId  string              `json:"client_msg_id,omitempty"`
	Typ          string              `json:"type"`
	Subtype      string              `json:"subtype,omitempty"`
	Text         string              `json:"text"`
	User         string              `json:"user"`
	Ts           string              `json:"ts"`
	ThreadTs     string              `json:"thread_ts,omitempty"`
	ParentUserId string              `json:"parent_user_id,omitempty"`
	Username     string              `json:"username,omitempty"`
	BotId        string              `json:"bot_id,omitempty"`
	Team         string              `json:"team,omitempty"`
	UserTeam     string              `json:"user_team,omitempty"`
	SourceTeam   string              `json:"source_team,omitempty"`
	UserProfile  *messageUserProfile `json:"user_profile,omitempty"`
	Attachments  []messageAttachment `json:"attachments,omitempty"`
	Blocks       []interface{}       `json:"blocks,omitempty"` // TODO: Use messageBlock
	Reactions    []messageReaction   `json:"reactions,omitempty"`
	Edited       *messageEdited      `json:"edited,omitempty"`
	Icons        *messageIcons       `json:"icons,omitempty"`
	Files        []messageFile       `json:"files,omitempty"`
	Root         *message            `json:"root,omitempty"`
	DisplayAsBot bool                `json:"display_as_bot,omitempty"`
	Upload       bool                `json:"upload,omitempty"`
	// if true, the message user the same as the previous one
	Trail bool `json:"-"`
}

type messageFile struct {
	Id                 string `json:"id"`
	Created            int64  `json:"created"`
	Timestamp          int64  `json:"timestamp"`
	Name               string `json:"name"`
	Title              string `json:"title"`
	Mimetype           string `json:"mimetype"`
	Filetype           string `json:"filetype"`
	PrettyType         string `json:"pretty_type"`
	User               string `json:"user"`
	Editable           bool   `json:"editable"`
	Size               int64  `json:"size"`
	Mode               string `json:"mode"`
	IsExternal         bool   `json:"is_external"`
	ExternalType       string `json:"external_type"`
	IsPublic           bool   `json:"is_public"`
	PublicUrlShared    bool   `json:"public_url_shared"`
	DisplayAsBot       bool   `json:"display_as_bot"`
	Username           string `json:"username"`
	UrlPrivate         string `json:"url_private"`
	UrlPrivateDownload string `json:"url_private_download"`
	Thumb64            string `json:"thumb_64,omitempty"`
	Thumb80            string `json:"thumb_80,omitempty"`
	Thumb160           string `json:"thumb_160,omitempty"`
	Thumb360           string `json:"thumb_360,omitempty"`
	Thumb360W          int64  `json:"thumb_360_w,omitempty"`
	Thumb360H          int64  `json:"thumb_360_h,omitempty"`
	Thumb480           string `json:"thumb_480,omitempty"`
	Thumb480W          int64  `json:"thumb_480_w,omitempty"`
	Thumb480H          int64  `json:"thumb_480_h,omitempty"`
	Thumb720           string `json:"thumb_720,omitempty"`
	Thumb720W          int64  `json:"thumb_720_w,omitempty"`
	Thumb720H          int64  `json:"thumb_720_h,omitempty"`
	Thumb800           string `json:"thumb_800,omitempty"`
	Thumb800W          int64  `json:"thumb_800_w,omitempty"`
	Thumb800H          int64  `json:"thumb_800_h,omitempty"`
	Thumb960           string `json:"thumb_960,omitempty"`
	Thumb960W          int64  `json:"thumb_960_w,omitempty"`
	Thumb960H          int64  `json:"thumb_960_h,omitempty"`
	Thumb1024          string `json:"thumb_1024,omitempty"`
	Thumb1024W         int64  `json:"thumb_1024_w,omitempty"`
	Thumb1024H         int64  `json:"thumb_1024_h,omitempty"`
	Thumb360Gif        string `json:"thumb_360_gif,omitempty"`
	Thumb480Gif        string `json:"thumb_480_gif,omitempty"`
	DeanimateGif       string `json:"deanimate_gif,omitempty"`
	ThumbTiny          string `json:"thumb_tiny,omitempty"`
	OriginalW          int64  `json:"original_w,omitempty"`
	OriginalH          int64  `json:"original_h,omitempty"`
	ThumbVideo         string `json:"thumb_video,omitempty"`
	Permalink          string `json:"permalink"`
	PermalinkPublic    string `json:"permalink_public"`
	EditLink           string `json:"edit_link,omitempty"`
	IsStarred          bool   `json:"is_starred"`
	HasRichPreview     bool   `json:"has_rich_preview"`
}

type messageIcons struct {
	Image48 string `json:"image_48"`
}

type messageEdited struct {
	User string `json:"user"`
	Ts   string `json:"ts"`
}

type messageUserProfile struct {
	AvatarHash        string `json:"avatar_hash"`
	Image72           string `json:"image72"`
	FirstName         string `json:"first_name"`
	RealName          string `json:"real_name"`
	DisplayName       string `json:"display_name"`
	Team              string `json:"team"`
	Name              string `json:"name"`
	IsRestricted      bool   `json:"is_restricted"`
	IsUltraRestricted bool   `json:"is_ultra_restricted"`
}

type messageBlock struct {
	Typ      string                `json:"type"`
	Elements []messageBlockElement `json:"elements"`
}

type messageBlockElement struct {
	Typ       string `json:"type"`
	Name      string `json:"name"`       // for type = "emoji"
	Text      string `json:"text"`       // for type = "text"
	ChannelId string `json:"channel_id"` // for type = "channel"
}

type messageAttachment struct {
	ServiceName     string `json:"service_name,omitempty"`
	AuthorIcon      string `json:"author_icon,omitempty"`
	AuthorName      string `json:"author_name,omitempty"`
	AuthorSubname   string `json:"author_subname,omitempty"`
	Title           string `json:"title,omitempty"`
	TitleLink       string `json:"title_link,omitempty"`
	Text            string `json:"text,omitempty"`
	Fallback        string `json:"fallback,omitempty"`
	ThumbUrl        string `json:"thumb_url,omitempty"`
	FromUrl         string `json:"from_url,omitempty"`
	ThumbWidth      int    `json:"thumb_width,omitempty"`
	ThumbHeight     int    `json:"thumb_height,omitempty"`
	ServiceIcon     string `json:"service_icon,omitempty"`
	Id              int    `json:"id"`
	OriginalUrl     string `json:"original_url,omitempty"`
	VideoHtml       string `json:"video_html,omitempty"`
	VideoHtmlWidth  int    `json:"video_html_width,omitempty"`
	VideoHtmlHeight int    `json:"video_html_height,omitempty"`
	Footer          string `json:"footer,omitempty"`
	FooterIcon      string `json:"footer_icon,omitempty"`
}

type messageReaction struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
	Count int      `json:"count"`
}

func readMessages(msgJsonPath string, threadMap msgThreadMap) ([]message, error) {
	content, err := ioutil.ReadFile(msgJsonPath)
	if err != nil {
		return nil, err
	}
	var msgs []message
	err = json.Unmarshal(content, &msgs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %s", msgJsonPath, err)
	}
	results := make([]message, 0)
	for i := range msgs {
		if !visibleMsg(&msgs[i]) {
			continue
		}
		rootMsgOfThread := msgs[i].ThreadTs == msgs[i].Ts
		if msgs[i].ThreadTs == "" || rootMsgOfThread ||
			msgs[i].Subtype == "thread_broadcast" ||
			msgs[i].Subtype == "bot_message" ||
			msgs[i].Subtype == "slackbot_response" {
			results = append(results, msgs[i])
		}
		if msgs[i].ThreadTs != "" {
			if rootMsgOfThread {
				threadTs := msgs[i].ThreadTs
				replies := threadMap[msgs[i].ThreadTs]
				for j := 0; j < len(replies); { // remove root message(s)
					if replies[j].Ts == threadTs {
						replies = append(replies[:j], replies[j+1:]...)
						continue
					}
					j++
				}
				threadMap[msgs[i].ThreadTs] = append([]*message{&msgs[i]}, replies...)
			} else {
				threadMap[msgs[i].ThreadTs] = append(threadMap[msgs[i].ThreadTs], &msgs[i])
			}
		}
	}
	return results, nil
}

type config struct {
	EditedSuffix string   `json:"edited_suffix"`
	Channels     []string `json:"channels"`
	EmojiJson    string   `json:"emoji_json"`
}

func readConfig(configPath string) (*config, error) {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg config
	err = json.Unmarshal(content, &cfg)
	return &cfg, err
}

func readEmojiJson(cfg *config, configJsonPath string) map[string]string {
	var emojis map[string]string
	if cfg.EmojiJson == "" {
		return emojis
	}
	emojiJsonPath := filepath.Join(filepath.Dir(configJsonPath), cfg.EmojiJson)
	content, err := ioutil.ReadFile(emojiJsonPath)
	if err != nil {
		return emojis
	}

	json.Unmarshal(content, &emojis)

	return emojis
}

type user struct {
	Id                string      `json:"id"`
	TeamId            string      `json:"team_id"`
	Name              string      `json:"name"`
	Deleted           bool        `json:"deleted"`
	Color             string      `json:"color"`
	RealName          string      `json:"real_name"`
	Tz                string      `json:"tz"`
	TzLabel           string      `json:"tz_label"`
	TzOffset          int         `json:"tz_offset"` // tzOffset / 60 / 60 = [-+] hour
	Profile           userProfile `json:"profile"`
	IsAdmin           bool        `json:"is_admin"`
	IsOwner           bool        `json:"is_owner"`
	IsPrimaryOwner    bool        `json:"is_primary_owner"`
	IsRestricted      bool        `json:"is_restricted"`
	IsUltraRestricted bool        `json:"is_ultra_restricted"`
	IsBot             bool        `json:"is_bot"`
	IsAppUser         bool        `json:"is_app_user"`
	Updated           int64       `json:"updated"`
}

type userProfile struct {
	Title                 string      `json:"title"`
	Phone                 string      `json:"phone"`
	Skype                 string      `json:"skype"`
	RealName              string      `json:"real_name"`
	RealNameNormalized    string      `json:"real_name_normalized"`
	DisplayName           string      `json:"display_name"`
	DisplayNameNormalized string      `json:"display_name_normalized"`
	Fields                interface{} `json:"fields"` // TODO ???
	StatusText            string      `json:"status_text"`
	StatusEmoji           string      `json:"status_emoji"`
	StatusExpiration      int64       `json:"status_expiration"`
	AvatarHash            string      `json:"avatar_hash"`
	FirstName             string      `json:"first_name"`
	LastName              string      `json:"last_name"`
	Image24               string      `json:"image_24"`
	Image32               string      `json:"image_32"`
	Image48               string      `json:"image_48"`
	Image72               string      `json:"image_72"`
	Image192              string      `json:"image_192"`
	Image512              string      `json:"image_512"`
	StatusTextCanonical   string      `json:"status_text_canonical"`
	Team                  string      `json:"team"`
	BotId                 string      `json:"bot_id"`
}

func readUsers(usersJsonPath string) ([]user, map[string]*user, error) {
	content, err := ioutil.ReadFile(usersJsonPath)
	if err != nil {
		return nil, nil, err
	}
	var users []user
	err = json.Unmarshal(content, &users)
	userMap := make(map[string]*user, len(users))
	for i := range users {
		userMap[users[i].Id] = &users[i]
		if users[i].Profile.BotId != "" {
			userMap[users[i].Profile.BotId] = &users[i]
		}
	}
	return users, userMap, err
}

type channel struct {
	Id         string         `json:"id"`
	Name       string         `json:"name"`
	Created    int64          `json:"created"`
	Creator    string         `json:"creator"`
	IsArchived bool           `json:"is_archived"`
	IsGeneral  bool           `json:"is_general"`
	Members    []string       `json:"members"`
	Pins       []channelPin   `json:"pins"`
	Topic      channelTopic   `json:"topic"`
	Purpose    channelPurpose `json:"purpose"`
}

type channelPin struct {
	Id      string `json:"id"`
	Typ     string `json:"type"`
	Created int64  `json:"created"`
	User    string `json:"user"`
	Owner   string `json:"owner"`
}

type channelTopic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

type channelPurpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

func readChannels(channelsJsonPath string, cfgChannels []string) ([]channel, map[string]*channel, error) {
	content, err := ioutil.ReadFile(channelsJsonPath)
	if err != nil {
		return nil, nil, err
	}
	var channels []channel
	err = json.Unmarshal(content, &channels)
	channels = filterChannel(channels, cfgChannels)
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	channelMap := make(map[string]*channel, len(channels))
	for i := range channels {
		channelMap[channels[i].Id] = &channels[i]
	}
	return channels, channelMap, err
}

func filterChannel(channels []channel, cfgChannels []string) []channel {
	newChannels := make([]channel, 0, len(channels))
	for i := range cfgChannels {
		if cfgChannels[i] == "*" {
			return channels
		}
		for j := range channels {
			if cfgChannels[i] == channels[j].Name {
				newChannels = append(newChannels, channels[j])
				break
			}
		}
	}
	return newChannels
}

var reToken = regexp.MustCompile(`\?t=xoxe-[-a-f0-9]+$`)

func (m *message) removeTokenFromURLs() {
	for i := range m.Files {
		m.Files[i].UrlPrivate = reToken.ReplaceAllLiteralString(m.Files[i].UrlPrivate, "")
		m.Files[i].UrlPrivateDownload = reToken.ReplaceAllLiteralString(m.Files[i].UrlPrivateDownload, "")
		m.Files[i].Thumb64 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb64, "")
		m.Files[i].Thumb80 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb80, "")
		m.Files[i].Thumb160 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb160, "")
		m.Files[i].Thumb360 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb360, "")
		m.Files[i].Thumb480 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb480, "")
		m.Files[i].Thumb720 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb720, "")
		m.Files[i].Thumb800 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb800, "")
		m.Files[i].Thumb960 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb960, "")
		m.Files[i].Thumb1024 = reToken.ReplaceAllLiteralString(m.Files[i].Thumb1024, "")
		m.Files[i].Thumb360Gif = reToken.ReplaceAllLiteralString(m.Files[i].Thumb360Gif, "")
		m.Files[i].Thumb480Gif = reToken.ReplaceAllLiteralString(m.Files[i].Thumb480Gif, "")
		m.Files[i].DeanimateGif = reToken.ReplaceAllLiteralString(m.Files[i].DeanimateGif, "")
		m.Files[i].ThumbVideo = reToken.ReplaceAllLiteralString(m.Files[i].ThumbVideo, "")
	}
}

func (f *messageFile) TopLevelMimetype() string {
	i := strings.Index(f.Mimetype, "/")
	if i < 0 {
		return ""
	}
	return f.Mimetype[:i]
}

func (f *messageFile) OriginalFilePath() string {
	suffix := f.downloadURLsAndSuffixes()[f.UrlPrivate]
	return f.Id + "/" + url.PathEscape(f.downloadFilename(f.UrlPrivate, suffix))
}

func (f *messageFile) ThumbImagePath() string {
	if f.Thumb1024 != "" {
		suffix := f.downloadURLsAndSuffixes()[f.Thumb1024]
		return f.Id + "/" + url.PathEscape(f.downloadFilename(f.Thumb1024, suffix))
	}
	return f.OriginalFilePath()
}

func (f *messageFile) ThumbImageWidth() int64 {
	if f.Thumb1024 != "" {
		return f.Thumb1024W
	}
	return f.OriginalW
}

func (f *messageFile) ThumbImageHeight() int64 {
	if f.Thumb1024 != "" {
		return f.Thumb1024H
	}
	return f.OriginalH
}

func (f *messageFile) ThumbVideoPath() string {
	suffix := f.downloadURLsAndSuffixes()[f.ThumbVideo]
	return f.Id + "/" + url.PathEscape(f.downloadFilename(f.ThumbVideo, suffix))
}
