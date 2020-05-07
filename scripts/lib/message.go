package slacklog

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Messages is an array of `*Message`.
type Messages []*Message

// Sort sorts messages by `Message.Ts` ascendant order.
func (msgs Messages) Sort() {
	sort.SliceStable(msgs, func(i, j int) bool {
		// must be the same digits, so no need to convert the timestamp to a number
		return msgs[i].Ts < msgs[j].Ts
	})
}

// MessagesMap is a map, maps MessageMonthKey to Messages.
type MessagesMap map[MessageMonthKey]Messages

// Keys returns all keys in the map.
func (mm MessagesMap) Keys() []MessageMonthKey {
	keys := make([]MessageMonthKey, 0, len(mm))
	for key := range mm {
		keys = append(keys, key)
	}
	return keys
}

// MessageTable : メッセージデータを保持する
// スレッドは投稿時刻からどのスレッドへの返信かが判断できるためThreadMapのキー
// はtsである。
// MsgsMapは月毎にメッセージを保持する。そのためキーは投稿月である。
// loadedFilesはすでに読み込んだファイルパスを保持する。
// loadedFilesは同じファイルを二度読むことを防ぐために用いている。
type MessageTable struct {
	// key: thread timestamp
	ThreadMap map[string]*Thread
	MsgsMap   MessagesMap
	// key: file path
	loadedFiles map[string]struct{}
}

// NewMessageTable : MessageTableを生成する。
// 他のテーブルと違い、メッセージファイルは量が多いため、NewMessageTable()実行
// 時には読み込まず、(*MessageTable).ReadLogDir()/(*MessageTable).ReadLogFile()
// 実行時に読み込ませる。
func NewMessageTable() *MessageTable {
	return &MessageTable{
		ThreadMap:   map[string]*Thread{},
		MsgsMap:     MessagesMap{},
		loadedFiles: map[string]struct{}{},
	}
}

// ReadLogDir : pathに指定したディレクトリに存在するJSON形式のメッセージデータ
// を読み込む。
// すでにそのディレクトリが読み込み済みの場合は処理をスキップする。
// visibleOnlyがtrueである場合は特定のサブタイプを持つメッセージのみをmsgMapに登録する。
func (m *MessageTable) ReadLogDir(path string, visibleOnly bool) error {
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
	for _, name := range names {
		if err := m.ReadLogFile(filepath.Join(path, name), visibleOnly); err != nil {
			return err
		}
	}
	return nil
}

// "{year}-{month}-{day}.json"
var reMsgFilename = regexp.MustCompile(`^(\d{4})-(\d{2})-\d{2}\.json$`)

// ReadLogFile : pathに指定したJSON形式のメッセージデータを読み込む。
// すでにそのファイルが読み込み済みの場合は処理をスキップする。
// readAllMessagesがfalseである場合は特定のサブタイプを持つメッセージのみをmsgMapに登録する。
func (m *MessageTable) ReadLogFile(path string, readAllMessages bool) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, ok := m.loadedFiles[path]; ok {
		return nil
	}

	match := reMsgFilename.FindStringSubmatch(filepath.Base(path))
	if len(match) == 0 {
		fmt.Fprintf(os.Stderr, "[warning] skipping %s ...\n", path)
		return nil
	}

	var msgs Messages
	err = ReadFileAsJSON(path, &msgs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", path, err)
	}

	// assort messages, visible and threaded.
	var visibleMsgs Messages
	for _, msg := range msgs {
		if readAllMessages && !msg.IsVisible() {
			continue
		}
		threadTs := msg.ThreadTs
		if threadTs != "" {
			thread, ok := m.ThreadMap[threadTs]
			if !ok {
				thread = &Thread{}
				m.ThreadMap[threadTs] = thread
			}
			thread.Put(msg)
		}

		if !readAllMessages &&
			threadTs != "" && !msg.IsRootOfThread() &&
			msg.Subtype != "thread_broadcast" &&
			msg.Subtype != "bot_message" &&
			msg.Subtype != "slackbot_response" {
			continue
		}

		visibleMsgs = append(visibleMsgs, msg)
	}

	key, err := NewMessageMonthKey(match[1], match[2])
	if err != nil {
		return err
	}
	if len(visibleMsgs) != 0 {
		m.MsgsMap[key] = append(m.MsgsMap[key], visibleMsgs...)
	}

	for _, msgs := range m.MsgsMap {
		msgs.Sort()
		var lastUser string
		for _, msg := range msgs {
			if lastUser == msg.User {
				msg.Trail = true
			} else {
				lastUser = msg.User
			}
		}
	}

	// loaded marker
	m.loadedFiles[path] = struct{}{}
	return nil
}

type MessageMonthKey struct {
	year  int
	month int
}

func NewMessageMonthKey(year, month string) (MessageMonthKey, error) {
	y, err := strconv.Atoi(year)
	if err != nil {
		return MessageMonthKey{}, err
	}
	m, err := strconv.Atoi(month)
	if err != nil {
		return MessageMonthKey{}, err
	}
	return MessageMonthKey{year: y, month: m}, nil
}

func (k MessageMonthKey) Next() MessageMonthKey {
	if k.month >= 12 {
		return MessageMonthKey{year: k.year + 1, month: 1}
	}
	return MessageMonthKey{year: k.year, month: k.month + 1}
}

func (k MessageMonthKey) Prev() MessageMonthKey {
	if k.month <= 1 {
		return MessageMonthKey{year: k.year - 1, month: 12}
	}
	return MessageMonthKey{year: k.year, month: k.month - 1}
}

func (k MessageMonthKey) Year() string {
	return fmt.Sprintf("%4d", k.year)
}

func (k MessageMonthKey) Month() string {
	return fmt.Sprintf("%02d", k.month)
}

func (k MessageMonthKey) NextYear() string {
	if k.month >= 12 {
		return fmt.Sprintf("%4d", k.year+1)
	}
	return fmt.Sprintf("%4d", k.year)
}

func (k MessageMonthKey) NextMonth() string {
	if k.month >= 12 {
		return "01"
	}
	return fmt.Sprintf("%02d", k.month+1)
}

func (k MessageMonthKey) PrevYear() string {
	if k.month <= 1 {
		return fmt.Sprintf("%4d", k.year-1)
	}
	return fmt.Sprintf("%4d", k.year)
}

func (k MessageMonthKey) PrevMonth() string {
	if k.month <= 1 {
		return "12"
	}
	return fmt.Sprintf("%02d", k.month-1)
}

// Message : メッセージ
// エクスポートしたYYYY-MM-DD.jsonの中身を保持する。
// https://slack.com/intl/ja-jp/help/articles/220556107-Slack-%E3%81%8B%E3%82%89%E3%82%A8%E3%82%AF%E3%82%B9%E3%83%9D%E3%83%BC%E3%83%88%E3%81%97%E3%81%9F%E3%83%87%E3%83%BC%E3%82%BF%E3%81%AE%E8%AA%AD%E3%81%BF%E6%96%B9
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

// IsVisible : 表示すべきメッセージ種別かを判定する。
// 例えばchannel_joinなどは投稿された出力する必要がないため、falseを返す。
func (m Message) IsVisible() bool {
	return m.Subtype == "" ||
		m.Subtype == "bot_message" ||
		m.Subtype == "slackbot_response" ||
		m.Subtype == "thread_broadcast"
}

// IsRootOfThread : メッセージがスレッドの最初のメッセージであるかを判定する。
func (m Message) IsRootOfThread() bool {
	return m.Ts == m.ThreadTs
}

var reToken = regexp.MustCompile(`\?t=xoxe-[-a-f0-9]+$`)

func removeToken(s string) string {
	return reToken.ReplaceAllLiteralString(s, "")
}

func (m *Message) RemoveTokenFromURLs() {
	for i, f := range m.Files {
		f.URLPrivate = removeToken(f.URLPrivate)
		f.URLPrivateDownload = removeToken(f.URLPrivateDownload)
		f.Thumb64 = removeToken(f.Thumb64)
		f.Thumb80 = removeToken(f.Thumb80)
		f.Thumb160 = removeToken(f.Thumb160)
		f.Thumb360 = removeToken(f.Thumb360)
		f.Thumb480 = removeToken(f.Thumb480)
		f.Thumb720 = removeToken(f.Thumb720)
		f.Thumb800 = removeToken(f.Thumb800)
		f.Thumb960 = removeToken(f.Thumb960)
		f.Thumb1024 = removeToken(f.Thumb1024)
		f.Thumb360Gif = removeToken(f.Thumb360Gif)
		f.Thumb480Gif = removeToken(f.Thumb480Gif)
		f.DeanimateGif = removeToken(f.DeanimateGif)
		f.ThumbVideo = removeToken(f.ThumbVideo)
		m.Files[i] = f
	}
}

// MessageFile :
// エクスポートしたYYYY-MM-DD.jsonの中身を保持する
// https://slack.com/intl/ja-jp/help/articles/220556107-Slack-%E3%81%8B%E3%82%89%E3%82%A8%E3%82%AF%E3%82%B9%E3%83%9D%E3%83%BC%E3%83%88%E3%81%97%E3%81%9F%E3%83%87%E3%83%BC%E3%82%BF%E3%81%AE%E8%AA%AD%E3%81%BF%E6%96%B9
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
	suffix := f.DownloadURLsAndSuffixes()[f.URLPrivate]
	return f.ID + "/" + url.PathEscape(f.DownloadFilename(f.URLPrivate, suffix))
}

func (f *MessageFile) ThumbImagePath() string {
	if f.Thumb1024 != "" {
		suffix := f.DownloadURLsAndSuffixes()[f.Thumb1024]
		return f.ID + "/" + url.PathEscape(f.DownloadFilename(f.Thumb1024, suffix))
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
	suffix := f.DownloadURLsAndSuffixes()[f.ThumbVideo]
	return f.ID + "/" + url.PathEscape(f.DownloadFilename(f.ThumbVideo, suffix))
}

func (f *MessageFile) DownloadURLsAndSuffixes() map[string]string {
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

var filenameReplacer = strings.NewReplacer(
	`\`, "_",
	"/", "_",
	":", "_",
	"*", "_",
	"?", "_",
	`"`, "_",
	"<", "_",
	">", "_",
	"|", "_",
)

func (f *MessageFile) DownloadFilename(url, suffix string) string {
	ext := filepath.Ext(url)
	nameExt := filepath.Ext(f.Name)
	name := f.Name[:len(f.Name)-len(ext)]
	if ext == "" {
		ext = nameExt
		if ext == "" {
			ext = filetypeToExtension[f.Filetype]
		}
	}

	filename := filenameReplacer.Replace(name + suffix + ext)

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
