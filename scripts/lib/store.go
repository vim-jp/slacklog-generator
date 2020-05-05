package slacklog

import (
	"fmt"
	"os"
	"path/filepath"
)

// LogStore : ログデータを各種テーブルを介して取得するための構造体。
// MessageTableはチャンネル毎に用意しているためmtsはチャンネルIDをキーとするmap
// となっている。
type LogStore struct {
	path string
	ut   *UserTable
	ct   *ChannelTable
	et   *EmojiTable
	// key: channel ID
	mts map[string]*MessageTable
}

// NewLogStore : 各テーブルを生成して、LogStoreを生成する。
func NewLogStore(dirPath string, cfg *Config) (*LogStore, error) {
	ut, err := NewUserTable(filepath.Join(dirPath, "users.json"))
	if err != nil {
		return nil, err
	}

	ct, err := NewChannelTable(filepath.Join(dirPath, "channels.json"), cfg.Channels)
	if err != nil {
		return nil, err
	}

	et, err := NewEmojiTable(filepath.Join(dirPath, cfg.EmojiJSONPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// EmojiTable is not required, so if the file just doesn't exist, continue
		// processing.
	}

	mts := make(map[string]*MessageTable, len(ct.Channels))
	for _, ch := range ct.Channels {
		mts[ch.ID] = NewMessageTable()
	}

	return &LogStore{
		path: dirPath,
		ut:   ut,
		ct:   ct,
		et:   et,
		mts:  mts,
	}, nil
}

func (s *LogStore) GetChannels() []Channel {
	return s.ct.Channels
}

func (s *LogStore) HasNextMonth(channelID string, key MessageMonthKey) bool {
	if mt, ok := s.mts[channelID]; ok && mt != nil {
		_, ok := mt.MsgsMap[key.Next()]
		return ok
	}
	return false
}

func (s *LogStore) HasPrevMonth(channelID string, key MessageMonthKey) bool {
	if mt, ok := s.mts[channelID]; ok && mt != nil {
		_, ok := mt.MsgsMap[key.Prev()]
		return ok
	}
	return false
}

func (s *LogStore) GetMessagesPerMonth(channelID string) (MessagesMap, error) {
	mt, ok := s.mts[channelID]
	if !ok {
		return nil, fmt.Errorf("not found channel: id=%s", channelID)
	}
	if err := mt.ReadLogDir(filepath.Join(s.path, channelID), true); err != nil {
		return nil, err
	}

	return mt.MsgsMap, nil
}

func (s *LogStore) GetUserByID(userID string) (*User, bool) {
	u, ok := s.ut.UserMap[userID]
	return u, ok
}

func (s *LogStore) GetDisplayNameByUserID(userID string) string {
	if user, ok := s.ut.UserMap[userID]; ok {
		if user.Profile.RealName != "" {
			return user.Profile.RealName
		}
		if user.Profile.DisplayName != "" {
			return user.Profile.DisplayName
		}
	}
	return ""
}

func (s *LogStore) GetDisplayNameMap() map[string]string {
	ret := make(map[string]string, len(s.ut.UserMap))
	for id, u := range s.ut.UserMap {
		ret[id] = s.GetDisplayNameByUserID(u.ID)
	}
	return ret
}

func (s *LogStore) GetEmojiMap() map[string]string {
	return s.et.NameToExt
}

func (s *LogStore) GetThread(channelID, ts string) (*Thread, bool) {
	mt, ok := s.mts[channelID]
	if !ok {
		return nil, false
	}
	if t, ok := mt.ThreadMap[ts]; ok {
		return t, true
	}
	return nil, false
}
