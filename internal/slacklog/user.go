package slacklog

// UserTable : ユーザデータを保持する
// UsersもUserMapも保持するユーザデータは同じで、UserMapはユーザIDをキーとする
// mapとなっている。
// ユースケースに応じてUsersとUserMapは使い分ける。
type UserTable struct {
	Users []User
	// key: user ID
	UserMap map[string]*User
}

// NewUserTable : pathに指定したJSON形式のユーザデータを読み込み、UserTableを生
// 成する。
func NewUserTable(path string) (*UserTable, error) {
	var users []User
	err := ReadFileAsJSON(path, &users)
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]*User, len(users))
	for i, u := range users {
		pu := &users[i]
		userMap[u.ID] = pu
		if u.Profile.BotID != "" {
			userMap[u.Profile.BotID] = pu
		}
	}
	return &UserTable{users, userMap}, nil
}

// User : ユーザ
// エクスポートしたuser.jsonの中身を保持する。
// 公式の情報は以下だがuser.jsonの解説までは書かれていない。
// https://slack.com/intl/ja-jp/help/articles/220556107-Slack-%E3%81%8B%E3%82%89%E3%82%A8%E3%82%AF%E3%82%B9%E3%83%9D%E3%83%BC%E3%83%88%E3%81%97%E3%81%9F%E3%83%87%E3%83%BC%E3%82%BF%E3%81%AE%E8%AA%AD%E3%81%BF%E6%96%B9
type User struct {
	ID                string      `json:"id"`
	TeamID            string      `json:"team_id"`
	Name              string      `json:"name"`
	Deleted           bool        `json:"deleted"`
	Color             string      `json:"color"`
	RealName          string      `json:"real_name"`
	TZ                string      `json:"tz"`
	TZLabel           string      `json:"tz_label"`
	TZOffset          int         `json:"tz_offset"` // tzOffset / 60 / 60 = [-+] hour
	Profile           UserProfile `json:"profile"`
	IsAdmin           bool        `json:"is_admin"`
	IsOwner           bool        `json:"is_owner"`
	IsPrimaryOwner    bool        `json:"is_primary_owner"`
	IsRestricted      bool        `json:"is_restricted"`
	IsUltraRestricted bool        `json:"is_ultra_restricted"`
	IsBot             bool        `json:"is_bot"`
	IsAppUser         bool        `json:"is_app_user"`
	Updated           int64       `json:"updated"`
}

// UserProfile : ユーザのプロファイル情報
// エクスポートしたuser.jsonの中身を保持する
// 公式の情報は以下だがuser.jsonの解説までは書かれていない。
// https://slack.com/intl/ja-jp/help/articles/220556107-Slack-%E3%81%8B%E3%82%89%E3%82%A8%E3%82%AF%E3%82%B9%E3%83%9D%E3%83%BC%E3%83%88%E3%81%97%E3%81%9F%E3%83%87%E3%83%BC%E3%82%BF%E3%81%AE%E8%AA%AD%E3%81%BF%E6%96%B9
type UserProfile struct {
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
	BotID                 string      `json:"bot_id"`

	// added for https://github.com/vim-jp/slacklog-generator/issues/69
	// 「unknown な JSON のフィールドがあったらエラーにする」
	AlwaysActive  bool   `json:"always_active"`
	ApiAppID      string `json:"api_app_id"`
	Image1024     string `json:"image_1024"`
	ImageOriginal string `json:"image_original"`
	IsCustomImage bool   `json:"is_custom_image"`
}
