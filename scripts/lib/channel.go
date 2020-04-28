package slacklog

import (
	"encoding/json"
	"io/ioutil"
	"sort"
)

type ChannelTable struct {
	l []Channel
	m map[string]*Channel
}

func NewChannelTable(path string, cfg []string) (*ChannelTable, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var channels []Channel
	err = json.Unmarshal(content, &channels)
	channels = FilterChannel(channels, cfg)
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	channelMap := make(map[string]*Channel, len(channels))
	for i := range channels {
		channelMap[channels[i].ID] = &channels[i]
	}
	return &ChannelTable{channels, channelMap}, err
}

func FilterChannel(channels []Channel, cfg []string) []Channel {
	newChannels := make([]Channel, 0, len(channels))
	for i := range cfg {
		if cfg[i] == "*" {
			return channels
		}
		for j := range channels {
			if cfg[i] == channels[j].Name {
				newChannels = append(newChannels, channels[j])
				break
			}
		}
	}
	return newChannels
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
