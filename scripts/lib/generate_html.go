package slacklog

import (
	"encoding/json"
	"io/ioutil"
	"sort"
)

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

type config struct {
	EditedSuffix string   `json:"edited_suffix"`
	Channels     []string `json:"channels"`
	EmojiJson    string   `json:"emoji_json"`
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
