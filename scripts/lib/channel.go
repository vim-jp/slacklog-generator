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
	if err := ReadFileAsJSON(path, &channels); err != nil {
		return nil, err
	}
	channels = FilterChannel(channels, whitelist)
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	channelMap := make(map[string]*Channel, len(channels))
	for i := range channels {
		channelMap[channels[i].ID] = &channels[i]
	}
	return &ChannelTable{channels, channelMap}, nil
}

// FilterChannel : whitelistに指定したチャンネル名に該当するチャンネルのみを返
// す。
// whitelistに'*'が含まれる場合はchannelをそのまま返す。
func FilterChannel(channels []Channel, whitelist []string) []Channel {
	newChannels := make([]Channel, 0, len(channels))
	for i := range whitelist {
		if whitelist[i] == "*" {
			return channels
		}
		for j := range channels {
			if whitelist[i] == channels[j].Name {
				newChannels = append(newChannels, channels[j])
				break
			}
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

type ChannelPin struct {
	ID      string `json:"id"`
	Typ     string `json:"type"`
	Created int64  `json:"created"`
	User    string `json:"user"`
	Owner   string `json:"owner"`
}

type ChannelTopic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

type ChannelPurpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}
