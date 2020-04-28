package slacklog

import (
	"path/filepath"
)

type LogStore struct {
	path string
	ut   *UserTable
	ct   *ChannelTable
	et   *EmojiTable
	mt   *MessageTable
}

func NewLogStore(dirPath string, cfg *Config) (*LogStore, error) {
	ut, err := NewUserTable(filepath.Join(dirPath, "users.json"))
	if err != nil {
		return nil, err
	}

	ct, err := NewChannelTable(filepath.Join(dirPath, "channels.json"), cfg.Channels)
	if err != nil {
		return nil, err
	}

	et := NewEmojiTable(filepath.Join(dirPath, cfg.EmojiJson))

	mt := NewMessageTable()

	return &LogStore{
		path: dirPath,
		ut:   ut,
		ct:   ct,
		et:   et,
		mt:   mt,
	}, nil
}

func (s *LogStore) GetChannels() []Channel {
	return s.ct.l
}

func (s *LogStore) HasNextMonth(msgsPerMonth MessagesPerMonth) bool {
	_, ok := s.mt.msgsMap[msgsPerMonth.NextKey()]
	return ok
}

func (s *LogStore) HasPrevMonth(msgsPerMonth MessagesPerMonth) bool {
	_, ok := s.mt.msgsMap[msgsPerMonth.PrevKey()]
	return ok
}

func (s *LogStore) GetMessagesPerMonth(channelID string) ([]MessagesPerMonth, error) {
	if err := s.mt.ReadLogDir(filepath.Join(s.path, channelID)); err != nil {
		return nil, err
	}

	msgs := s.mt.msgsMap
	ret := make([]MessagesPerMonth, len(msgs))
	i := 0
	for _, msgsPerMonth := range msgs {
		ret[i] = msgsPerMonth
		i++
	}

	return ret, nil
}

func (s *LogStore) GetUserByID(userID string) (*User, bool) {
	u, ok := s.ut.m[userID]
	return u, ok
}

func (s *LogStore) GetDisplayNameByUserID(userID string) string {
	if user, ok := s.ut.m[userID]; ok {
		if user.Profile.RealName != "" {
			return user.Profile.RealName
		}
		if user.Profile.DisplayName != "" {
			return user.Profile.DisplayName
		}
	}
	return ""
}

func (s *LogStore) GetUserNameMap() map[string]string {
	ret := make(map[string]string, len(s.ut.m))
	for id, u := range s.ut.m {
		ret[id] = u.Name
	}
	return ret
}

func (s *LogStore) GetEmojiMap() map[string]string {
	return s.et.m
}

func (s *LogStore) GetThread(ts string) (*Thread, bool) {
	if t, ok := s.mt.threadMap[ts]; ok {
		return &t, true
	}
	return nil, false
}
