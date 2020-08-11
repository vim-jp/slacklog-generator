package slacklog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

// Messages is an array of `*Message`.
type Messages []*Message

// Sort sorts messages by `Message.Ts` ascendant order.
func (msgs Messages) Sort() {
	sort.SliceStable(msgs, func(i, j int) bool {
		// must be the same digits, so no need to convert the timestamp to a number
		return msgs[i].Timestamp < msgs[j].Timestamp
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
	// key: timestamp
	AllMessageMap map[string]*Message
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
		AllMessageMap: map[string]*Message{},
		ThreadMap:     map[string]*Thread{},
		MsgsMap:       MessagesMap{},
		loadedFiles:   map[string]struct{}{},
	}
}

// ReadLogDir : pathに指定したディレクトリに存在するJSON形式のメッセージデータ
// を読み込む。
// すでにそのディレクトリが読み込み済みの場合は処理をスキップする。
// デフォルトでは特定のサブタイプを持つメッセージのみをmsgMapに登録するが、
// readAllMessages が true である場合はすべてのメッセージを登録する。
func (m *MessageTable) ReadLogDir(path string, readAllMessages bool) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	// ReadLogFile()は日付順に処理する必要があり、そのためにファイル名でソート
	// している
	sort.Strings(names)
	for _, name := range names {
		if err := m.ReadLogFile(filepath.Join(path, name), readAllMessages); err != nil {
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
	err = ReadFileAsJSON(path, true, &msgs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", path, err)
	}

	// assort messages, visible and threaded.
	var visibleMsgs Messages
	for _, msg := range msgs {
		if !readAllMessages && !msg.isVisible() {
			continue
		}

		// スレッドに所属してるメッセージは ThreadMap へスレッド毎に分別しておく
		threadTs := msg.ThreadTimestamp
		if threadTs != "" {
			thread, ok := m.ThreadMap[threadTs]
			if !ok {
				thread = &Thread{}
				if rootMsg, ok := m.AllMessageMap[threadTs]; ok {
					thread.rootMsg = rootMsg
				}
				m.ThreadMap[threadTs] = thread
			}
			thread.Put(msg)
		}

		if !readAllMessages && msg.isThreadChild() {
			continue
		}

		m.AllMessageMap[msg.Timestamp] = msg

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

// MessageMonthKey is a key for messages.
type MessageMonthKey struct {
	year  int
	month int
}

// NewMessageMonthKey creates MessageMonthKey key from two strings: which
// represents year and month.
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

// Next gets a key for next month.
func (k MessageMonthKey) Next() MessageMonthKey {
	if k.month >= 12 {
		return MessageMonthKey{year: k.year + 1, month: 1}
	}
	return MessageMonthKey{year: k.year, month: k.month + 1}
}

// Prev gets a key for previous month.
func (k MessageMonthKey) Prev() MessageMonthKey {
	if k.month <= 1 {
		return MessageMonthKey{year: k.year - 1, month: 12}
	}
	return MessageMonthKey{year: k.year, month: k.month - 1}
}

// Year returns string represents year.
func (k MessageMonthKey) Year() string {
	return fmt.Sprintf("%4d", k.year)
}

// Month returns string represents month.
func (k MessageMonthKey) Month() string {
	return fmt.Sprintf("%02d", k.month)
}

// NextYear returns a string for next year.
func (k MessageMonthKey) NextYear() string {
	if k.month >= 12 {
		return fmt.Sprintf("%4d", k.year+1)
	}
	return fmt.Sprintf("%4d", k.year)
}

// NextMonth returns a string for next month.
func (k MessageMonthKey) NextMonth() string {
	if k.month >= 12 {
		return "01"
	}
	return fmt.Sprintf("%02d", k.month+1)
}

// PrevYear returns a string for previous year.
func (k MessageMonthKey) PrevYear() string {
	if k.month <= 1 {
		return fmt.Sprintf("%4d", k.year-1)
	}
	return fmt.Sprintf("%4d", k.year)
}

// PrevMonth returns a string for previous month.
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
	slack.Message

	// Trail shows the user of message is same as the previous one.
	// FIXME: 本来はココに書いてはいけない
	Trail bool `json:"-"`
}

// isVisible : 表示すべきメッセージ種別かを判定する。
// 例えばchannel_joinなどは投稿された出力する必要がないため、falseを返す。
func (m *Message) isVisible() bool {
	return m.SubType == "" ||
		m.SubType == "bot_message" ||
		m.SubType == "slackbot_response" ||
		m.SubType == "thread_broadcast"
}

// isBotMessage : メッセージがBotからの物かを判定する。
func (m *Message) isBotMessage() bool {
	return m.SubType == "bot_message" ||
		m.SubType == "slackbot_response"
}

// IsRootOfThread : メッセージがスレッドの最初のメッセージであるかを判定する。
func (m Message) IsRootOfThread() bool {
	return m.Timestamp == m.ThreadTimestamp
}

// isThreadChild returns true when a message should be shown in a thread only.
func (m *Message) isThreadChild() bool {
	return m.ThreadTimestamp != "" && m.Timestamp != m.ThreadTimestamp && m.SubType != "thread_broadcast"
}

var reToken = regexp.MustCompile(`\?t=xoxe-[-a-f0-9]+$`)

func removeToken(s string) string {
	return reToken.ReplaceAllLiteralString(s, "")
}

// RemoveTokenFromURLs removes the token from URLs in a message.
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

// RegulateFilename replaces unusable characters as filepath by '_'.
func RegulateFilename(s string) string {
	return filenameReplacer.Replace(s)
}

// MessageIcons represents icon for each message.
type MessageIcons struct {
	Image48 string `json:"image_48"`
}
