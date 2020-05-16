package slacklog

import (
	"sort"
)

// ChannelTable : チャンネルデータを保持する。
// channelsもchannelMapも保持するチャンネルデータは同じで、channelMapはチャンネ
// ルIDをキーとするmapとなっている。
// ユースケースに応じてchannelsとchannelMapは使い分ける。
type ChannelTable struct {
	Channels   []Channel
	ChannelMap map[string]*Channel
}

// NewChannelTable : pathに指定したJSON形式のチャンネルデータを読み込み、
// ChannelTable を生成する。
// whitelistに指定したチャンネル名のみを読み込む。
func NewChannelTable(path string, whitelist []string) (*ChannelTable, error) {
	var channels []Channel
	if err := ReadFileAsJSON(path, true, &channels); err != nil {
		return nil, err
	}
	channels = FilterChannel(channels, whitelist)
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	channelMap := make(map[string]*Channel, len(channels))
	for i, ch := range channels {
		channelMap[ch.ID] = &channels[i]
	}
	return &ChannelTable{
		Channels:   channels,
		ChannelMap: channelMap,
	}, nil
}

// FilterChannel : whitelistに指定したチャンネル名に該当するチャンネルのみを返
// す。
// whitelistに'*'が含まれる場合はchannelをそのまま返す。
func FilterChannel(channels []Channel, whitelist []string) []Channel {
	if len(whitelist) == 0 {
		return []Channel{}
	}
	allowed := map[string]struct{}{}
	for _, s := range whitelist {
		if s == "*" {
			return channels
		}
		allowed[s] = struct{}{}
	}
	newChannels := make([]Channel, 0, len(whitelist))
	for _, ch := range channels {
		_, ok := allowed[ch.Name]
		if ok {
			newChannels = append(newChannels, ch)
		}
	}
	return newChannels
}

// SortChannel sorts []Channel by name. It modify original slice.
func SortChannel(channels []Channel) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
}

// Channel represents channel object in Slack.
type Channel struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Created    int64          `json:"created"`
	Creator    string         `json:"creator"`
	IsArchived bool           `json:"is_archived"`
	IsGeneral  bool           `json:"is_general"`
	Members    []string       `json:"members"`
	Pins       []ChannelPin   `json:"pins"`
	Topic      ChannelTopic   `json:"topic"`
	Purpose    ChannelPurpose `json:"purpose"`
}

// ChannelPin represents a pinned message for a channel.
type ChannelPin struct {
	ID      string `json:"id"`
	Typ     string `json:"type"`
	Created int64  `json:"created"`
	User    string `json:"user"`
	Owner   string `json:"owner"`
}

// ChannelTopic represents topic of a channel.
type ChannelTopic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// ChannelPurpose represents puropse of a channel.
type ChannelPurpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}
