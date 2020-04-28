package slacklog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type MessageTable struct {
	threadMap map[string]*Thread
	msgsMap   map[string]*MessagesPerMonth
	readDir   map[string]struct{}
	readFile  map[string]struct{}
}

func NewMessageTable() *MessageTable {
	return &MessageTable{
		threadMap: map[string]*Thread{},
		msgsMap:   map[string]*MessagesPerMonth{},
		readDir:   map[string]struct{}{},
		readFile:  map[string]struct{}{},
	}
}

func (m *MessageTable) ReadLogDir(path string) error {
	if _, ok := m.readDir[path]; ok {
		return nil
	}

	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	sort.Strings(names)
	for i := range names {
		if err := m.ReadLogFile(filepath.Join(path, names[i])); err != nil {
			return err
		}
	}
	// read marker
	m.readDir[path] = struct{}{}
	return nil
}

func (m *MessageTable) ReadLogFile(path string) error {
	if _, ok := m.readFile[path]; ok {
		return nil
	}

	match := reMsgFilename.FindStringSubmatch(filepath.Base(path))
	if len(match) == 0 {
		fmt.Fprintf(os.Stderr, "[warning] skipping %s ...\n", path)
		return nil
	}
	key := match[1] + match[2]
	msgPerMonth, ok := m.msgsMap[key]
	if !ok {
		y, err := strconv.Atoi(match[1])
		if err != nil {
			return err
		}
		m, err := strconv.Atoi(match[2])
		if err != nil {
			return err
		}
		msgPerMonth = &MessagesPerMonth{year: y, month: m}
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var msgs []Message
	err = json.Unmarshal(content, &msgs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %s", path, err)
	}
	for i := range msgs {
		if !msgs[i].IsVisible() {
			continue
		}
		threadTs := msgs[i].ThreadTs
		if threadTs == "" || msgs[i].IsRootOfThread() ||
			msgs[i].Subtype == "thread_broadcast" ||
			msgs[i].Subtype == "bot_message" ||
			msgs[i].Subtype == "slackbot_response" {
			msgPerMonth.Messages = append(msgPerMonth.Messages, msgs[i])
			m.msgsMap[key] = msgPerMonth
		}
		if threadTs != "" {
			if m.threadMap[threadTs] == nil {
				m.threadMap[threadTs] = &Thread{}
			}
			if msgs[i].IsRootOfThread() {
				replies := m.threadMap[threadTs].msgs
				for j := 0; j < len(replies); { // remove root message(s)
					if replies[j].Ts == threadTs {
						replies = append(replies[:j], replies[j+1:]...)
						continue
					}
					j++
				}
				m.threadMap[threadTs].msgs = append([]Message{msgs[i]}, replies...)
			} else {
				m.threadMap[threadTs].msgs = append(m.threadMap[msgs[i].ThreadTs].msgs, msgs[i])
			}
		}
	}
	for key := range m.msgsMap {
		if len(m.msgsMap[key].Messages) == 0 {
			delete(m.msgsMap, key)
			continue
		}
		sort.SliceStable(m.msgsMap[key].Messages, func(i, j int) bool {
			// must be the same digits, so no need to convert the timestamp to a number
			return m.msgsMap[key].Messages[i].Ts < m.msgsMap[key].Messages[j].Ts
		})
		ms := m.msgsMap[key].Messages
		var lastUser string
		for i := range ms {
			if lastUser == ms[i].User {
				(&ms[i]).Trail = true
			} else {
				lastUser = ms[i].User
			}
		}
	}

	// read marker
	m.readFile[path] = struct{}{}
	return nil
}

type MessagesPerMonth struct {
	year     int
	month    int
	Messages []Message
}

func (m MessagesPerMonth) Year() string {
	return fmt.Sprintf("%4d", m.year)
}

func (m MessagesPerMonth) Month() string {
	return fmt.Sprintf("%02d", m.month)
}

func (m MessagesPerMonth) NextYear() string {
	if m.month >= 12 {
		return fmt.Sprintf("%4d", m.year+1)
	}
	return fmt.Sprintf("%4d", m.year)
}

func (m MessagesPerMonth) NextMonth() string {
	if m.month >= 12 {
		return "01"
	}
	return fmt.Sprintf("%02d", m.month+1)
}

func (m MessagesPerMonth) PrevYear() string {
	if m.month <= 1 {
		return fmt.Sprintf("%4d", m.year-1)
	}
	return fmt.Sprintf("%4d", m.year)
}

func (m MessagesPerMonth) PrevMonth() string {
	if m.month <= 1 {
		return "12"
	}
	return fmt.Sprintf("%02d", m.month-1)
}

func (m MessagesPerMonth) Key() string {
	return fmt.Sprintf("%4d%02d", m.year, m.month)
}

func (m MessagesPerMonth) NextKey() string {
	if m.month >= 12 {
		return fmt.Sprintf("%4d%02d", m.year+1, 1)
	}
	return fmt.Sprintf("%4d%02d", m.year, m.month+1)
}

func (m MessagesPerMonth) PrevKey() string {
	if m.month <= 1 {
		return fmt.Sprintf("%4d%02d", m.year-1, 12)
	}
	return fmt.Sprintf("%4d%02d", m.year, m.month-1)
}

type Message struct {
	ClientMsgID  string              `json:"client_msg_id,omitempty"`
	Typ          string              `json:"type"`
	Subtype      string              `json:"subtype,omitempty"`
	Text         string              `json:"text"`
	User         string              `json:"user"`
	Ts           string              `json:"ts"`
	ThreadTs     string              `json:"thread_ts,omitempty"`
	ParentUserID string              `json:"parent_user_id,omitempty"`
	Username     string              `json:"username,omitempty"`
	BotID        string              `json:"bot_id,omitempty"`
	Team         string              `json:"team,omitempty"`
	UserTeam     string              `json:"user_team,omitempty"`
	SourceTeam   string              `json:"source_team,omitempty"`
	UserProfile  *MessageUserProfile `json:"user_profile,omitempty"`
	Attachments  []MessageAttachment `json:"attachments,omitempty"`
	Blocks       []interface{}       `json:"blocks,omitempty"` // TODO: Use messageBlock
	Reactions    []MessageReaction   `json:"reactions,omitempty"`
	Edited       *MessageEdited      `json:"edited,omitempty"`
	Icons        *MessageIcons       `json:"icons,omitempty"`
	Files        []MessageFile       `json:"files,omitempty"`
	Root         *Message            `json:"root,omitempty"`
	DisplayAsBot bool                `json:"display_as_bot,omitempty"`
	Upload       bool                `json:"upload,omitempty"`
	// if true, the message user the same as the previous one
	Trail bool `json:"-"`
}

func (m Message) IsVisible() bool {
	return m.Subtype == "" ||
		m.Subtype == "bot_message" ||
		m.Subtype == "slackbot_response" ||
		m.Subtype == "thread_broadcast"
}

func (m Message) IsRootOfThread() bool {
	return m.Ts == m.ThreadTs
}

type MessageFile struct {
	ID                 string `json:"id"`
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
	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`
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

func (f *MessageFile) TopLevelMimetype() string {
	i := strings.Index(f.Mimetype, "/")
	if i < 0 {
		return ""
	}
	return f.Mimetype[:i]
}

func (f *MessageFile) OriginalFilePath() string {
	suffix := f.downloadURLsAndSuffixes()[f.URLPrivate]
	return f.ID + "/" + url.PathEscape(f.downloadFilename(f.URLPrivate, suffix))
}

func (f *MessageFile) ThumbImagePath() string {
	if f.Thumb1024 != "" {
		suffix := f.downloadURLsAndSuffixes()[f.Thumb1024]
		return f.ID + "/" + url.PathEscape(f.downloadFilename(f.Thumb1024, suffix))
	}
	return f.OriginalFilePath()
}

func (f *MessageFile) ThumbImageWidth() int64 {
	if f.Thumb1024 != "" {
		return f.Thumb1024W
	}
	return f.OriginalW
}

func (f *MessageFile) ThumbImageHeight() int64 {
	if f.Thumb1024 != "" {
		return f.Thumb1024H
	}
	return f.OriginalH
}

func (f *MessageFile) ThumbVideoPath() string {
	suffix := f.downloadURLsAndSuffixes()[f.ThumbVideo]
	return f.ID + "/" + url.PathEscape(f.downloadFilename(f.ThumbVideo, suffix))
}

func (f *MessageFile) downloadURLsAndSuffixes() map[string]string {
	return map[string]string{
		f.URLPrivate:   "",
		f.Thumb64:      "_64",
		f.Thumb80:      "_80",
		f.Thumb160:     "_160",
		f.Thumb360:     "_360",
		f.Thumb480:     "_480",
		f.Thumb720:     "_720",
		f.Thumb800:     "_800",
		f.Thumb960:     "_960",
		f.Thumb1024:    "_1024",
		f.Thumb360Gif:  "_360",
		f.Thumb480Gif:  "_480",
		f.DeanimateGif: "_deanimate_gif",
		f.ThumbVideo:   "_thumb_video",
	}
}

func (f *MessageFile) downloadFilename(url, suffix string) string {
	ext := filepath.Ext(url)
	nameExt := filepath.Ext(f.Name)
	name := f.Name[:len(f.Name)-len(ext)]
	if ext == "" {
		ext = nameExt
		if ext == "" {
			ext = filetypeToExtension[f.Filetype]
		}
	}

	filename := strings.ReplaceAll(name+suffix+ext, "/", "_")

	// XXX: Jekyll doesn't publish files that name starts with some characters
	if strings.HasPrefix(filename, "_") || strings.HasPrefix(filename, ".") {
		filename = "files" + filename
	}

	return filename
}

type MessageIcons struct {
	Image48 string `json:"image_48"`
}

type MessageEdited struct {
	User string `json:"user"`
	Ts   string `json:"ts"`
}

type MessageUserProfile struct {
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

type MessageBlock struct {
	Typ      string                `json:"type"`
	Elements []MessageBlockElement `json:"elements"`
}

type MessageBlockElement struct {
	Typ       string `json:"type"`
	Name      string `json:"name"`       // for type = "emoji"
	Text      string `json:"text"`       // for type = "text"
	ChannelID string `json:"channel_id"` // for type = "channel"
}

type MessageAttachment struct {
	ServiceName     string `json:"service_name,omitempty"`
	AuthorIcon      string `json:"author_icon,omitempty"`
	AuthorName      string `json:"author_name,omitempty"`
	AuthorSubname   string `json:"author_subname,omitempty"`
	Title           string `json:"title,omitempty"`
	TitleLink       string `json:"title_link,omitempty"`
	Text            string `json:"text,omitempty"`
	Fallback        string `json:"fallback,omitempty"`
	ThumbURL        string `json:"thumb_url,omitempty"`
	FromURL         string `json:"from_url,omitempty"`
	ThumbWidth      int    `json:"thumb_width,omitempty"`
	ThumbHeight     int    `json:"thumb_height,omitempty"`
	ServiceIcon     string `json:"service_icon,omitempty"`
	ID              int    `json:"id"`
	OriginalURL     string `json:"original_url,omitempty"`
	VideoHTML       string `json:"video_html,omitempty"`
	VideoHTMLWidth  int    `json:"video_html_width,omitempty"`
	VideoHTMLHeight int    `json:"video_html_height,omitempty"`
	Footer          string `json:"footer,omitempty"`
	FooterIcon      string `json:"footer_icon,omitempty"`
}

type MessageReaction struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
	Count int      `json:"count"`
}
