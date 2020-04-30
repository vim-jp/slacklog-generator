package subcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	slacklog "github.com/vim-jp/slacklog/lib"
)

func doConvertExportedLogs() error {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/main.go convert_exported_logs {indir} {outdir}")
		return nil
	}

	inDir := filepath.Clean(os.Args[2])
	outDir := filepath.Clean(os.Args[3])

	channels, _, err := readChannels(filepath.Join(inDir, "channels.json"), []string{"*"})
	if err != nil {
		return fmt.Errorf("could not read channels.json: %s", err)
	}

	if err := slacklog.Mkdir(outDir); err != nil {
		return fmt.Errorf("could not create %s directory: %s", outDir, err)
	}

	err = copyFile(filepath.Join(inDir, "channels.json"), filepath.Join(outDir, "channels.json"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(inDir, "users.json"), filepath.Join(outDir, "users.json"))
	if err != nil {
		return err
	}

	for _, channel := range channels {
		messages, err := ReadAllMessages(filepath.Join(inDir, channel.Name))
		if err != nil {
			return err
		}
		for _, message := range messages {
			message.UserProfile = nil
			message.RemoveTokenFromURLs()
		}
		channelDir := filepath.Join(outDir, channel.ID)
		if err := slacklog.Mkdir(channelDir); err != nil {
			return fmt.Errorf("could not create %s directory: %s", channelDir, err)
		}
		messagesPerDay := groupMessagesByDay(messages)
		for key := range messagesPerDay {
			err = writeMessages(filepath.Join(channelDir, key+".json"), messagesPerDay[key])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(from string, to string) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)

	return err
}

func readChannels(channelsJsonPath string, cfgChannels []string) ([]slacklog.Channel, map[string]*slacklog.Channel, error) {
	content, err := ioutil.ReadFile(channelsJsonPath)
	if err != nil {
		return nil, nil, err
	}
	var channels []slacklog.Channel
	err = json.Unmarshal(content, &channels)
	channels = slacklog.FilterChannel(channels, cfgChannels)
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	channelMap := make(map[string]*slacklog.Channel, len(channels))
	for i := range channels {
		channelMap[channels[i].ID] = &channels[i]
	}
	return channels, channelMap, err
}

func ReadAllMessages(inDir string) ([]*slacklog.Message, error) {
	dir, err := os.Open(inDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var messages []*slacklog.Message
	for i := range names {
		content, err := ioutil.ReadFile(filepath.Join(inDir, names[i]))
		if err != nil {
			return nil, err
		}
		var msgs []*slacklog.Message
		err = json.Unmarshal(content, &msgs)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msgs...)
	}
	return messages, nil
}

func groupMessagesByDay(messages []*slacklog.Message) map[string][]*slacklog.Message {
	messagesPerDay := map[string][]*slacklog.Message{}
	for i := range messages {
		time := slacklog.TsToDateTime(messages[i].Ts).Format("2006-01-02")
		messagesPerDay[time] = append(messagesPerDay[time], messages[i])
	}
	return messagesPerDay
}

func writeMessages(filename string, messages []*slacklog.Message) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(messages)
	if err != nil {
		return err
	}
	return nil
}
