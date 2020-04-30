package slacklog

import (
	"errors"
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

	et, err := NewEmojiTable(filepath.Join(dirPath, cfg.EmojiJson))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// EmojiTable is not required, so if the file just doesn't exist, continue
		// processing.
	}

	mts := make(map[string]*MessageTable, len(ct.channelMap))
	for channelID := range ct.channelMap {
		mts[channelID] = NewMessageTable()
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
	return s.ct.channels
}

func (s *LogStore) HasNextMonth(channelID string, msgsPerMonth MessagesPerMonth) bool {
	if mt, ok := s.mts[channelID]; ok && mt != nil {
		_, ok := mt.msgsMap[msgsPerMonth.NextKey()]
		return ok
	}
	return false
}

func (s *LogStore) HasPrevMonth(channelID string, msgsPerMonth MessagesPerMonth) bool {
	if mt, ok := s.mts[channelID]; ok && mt != nil {
		_, ok := mt.msgsMap[msgsPerMonth.PrevKey()]
		return ok
	}
	return false
}

func (s *LogStore) GetMessagesPerMonth(channelID string) ([]MessagesPerMonth, error) {
	mt, ok := s.mts[channelID]
	if !ok {
		return nil, errors.New("not found")
	}
	if err := mt.ReadLogDir(filepath.Join(s.path, channelID)); err != nil {
		return nil, err
	}

	msgs := mt.msgsMap
	ret := make([]MessagesPerMonth, len(msgs))
	i := 0
	for _, msgsPerMonth := range msgs {
		ret[i] = *msgsPerMonth
		i++
	}

	return ret, nil
}

func (s *LogStore) GetUserByID(userID string) (*User, bool) {
	u, ok := s.ut.userMap[userID]
	return u, ok
}

func (s *LogStore) GetDisplayNameByUserID(userID string) string {
	if user, ok := s.ut.userMap[userID]; ok {
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
	ret := make(map[string]string, len(s.ut.userMap))
	for id, u := range s.ut.userMap {
		ret[id] = s.GetDisplayNameByUserID(u.ID)
	}
	return ret
}

func (s *LogStore) GetEmojiMap() map[string]string {
	return s.et.urlMap
}

func (s *LogStore) GetThread(channelID, ts string) (*Thread, bool) {
	mt, ok := s.mts[channelID]
	if !ok {
		return nil, false
	}
	if t, ok := mt.threadMap[ts]; ok {
		return t, true
	}
	return nil, false
}
