package slacklog

import "github.com/slack-go/slack"

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
	err := ReadFileAsJSON(path, true, &users)
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
type User slack.User
