package slacklog

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

const gramN = 2

type Indexer struct {
	s              *LogStore
	gramIndex      messageIndex
	channelNumbers map[int]Channel
}

func NewIndexer(s *LogStore) *Indexer {
	return &Indexer{
		s:              s,
		gramIndex:      messageIndex{},
		channelNumbers: map[int]Channel{},
	}
}

func (idx *Indexer) Build() error {
	channelNumber := 0
	for _, c := range idx.s.GetChannels() {
		channelNumber++
		idx.channelNumbers[channelNumber] = c
		msgs, err := idx.s.GetAllMessages(c.ID)
		if err != nil {
			return err
		}
		for _, m := range msgs {
			runes := []rune(m.Text)
			textLen := len(runes)
			for i := range runes {
				for n := 1; n <= gramN; n++ {
					if textLen <= i+n-1 {
						break
					}
					key := string(runes[i : i+n])
					idx.gramIndex.Add(key, channelNumber, m, i)
				}
			}
		}
	}
	return nil
}

func (idx *Indexer) Output(outDir string) error {
	channelFilepath := filepath.Join(outDir, "channel")
	err := idx.writeChannelFile(channelFilepath, idx.channelNumbers)
	if err != nil {
		return err
	}

	for key, mPositions := range idx.gramIndex {
		s := outDir
		for _, u := range utf16.Encode([]rune(key)) {
			s = filepath.Join(s, fmt.Sprintf("%02x", u>>8), fmt.Sprintf("%02x", u&0xff))
		}
		err := idx.writeIndexFile(s+".index", mPositions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (idx *Indexer) writeChannelFile(path string, channelNumbers map[int]Channel) error {
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fw := bufio.NewWriter(f)
	for channelNumber, channel := range channelNumbers {
		_, err := fw.Write([]byte(fmt.Sprintf("%d\t%s\t%s\n", channelNumber, channel.ID, channel.Name)))
		if err != nil {
			return err
		}
	}

	err = fw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (idx *Indexer) writeIndexFile(path string, mPositions messagePositions) error {
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fw := bufio.NewWriter(f)
	for channelNumber, mposMap := range mPositions {
		_, err := fw.Write(vintBytes(channelNumber))
		if err != nil {
			return err
		}

		_, err = fw.Write(vintBytes(len(mposMap)))
		if err != nil {
			return err
		}

		for ts, positions := range mposMap {
			tsParts := strings.SplitN(ts, ".", 2)
			if len(tsParts) != 2 {
				channel := idx.channelNumbers[channelNumber]
				return fmt.Errorf("Invalid timestamp %s (%s)", ts, channel.ID)
			}

			tsSec, err := strconv.Atoi(tsParts[0])
			if err != nil {
				return err
			}
			err = binary.Write(fw, binary.BigEndian, uint32(tsSec))
			if err != nil {
				return err
			}

			tsMicrosec, err := strconv.Atoi(tsParts[1])
			if err != nil {
				return err
			}
			_, err = fw.Write(vintBytes(tsMicrosec))
			if err != nil {
				return err
			}

			if len(positions) == 0 {
				channel := idx.channelNumbers[channelNumber]
				return fmt.Errorf("Empty positions: %s: %s: %s", path, channel.ID, ts)
			}
			for _, pos := range positions {
				_, err = fw.Write(vintBytes(pos + 1))
				if err != nil {
					return err
				}
			}
			err = fw.WriteByte(0)
			if err != nil {
				return err
			}
		}
	}

	err = fw.Flush()
	if err != nil {
		return err
	}

	return nil
}

type messageIndex map[string]messagePositions

func (mi messageIndex) Add(key string, channelNumber int, mes *Message, pos int) {
	mp, ok := mi[key]
	if !ok {
		mp = messagePositions{}
		mi[key] = mp
	}
	mp.Add(channelNumber, mes, pos)
}

type messagePositions map[int]map[string][]int

func (mp messagePositions) Add(channelNumber int, mes *Message, pos int) {
	mposMap, ok := mp[channelNumber]
	if !ok {
		mposMap = map[string][]int{}
		mp[channelNumber] = mposMap
	}

	mposMap[mes.Timestamp] = append(mposMap[mes.Timestamp], pos)
}

func vintBytes(n int) []byte {
	if n == 0 {
		return []byte{0}
	}
	bytes := []byte{}
	cont := false
	for n != 0 {
		b := byte(n & 0b01111111)
		if cont {
			b |= 0b10000000
		}
		bytes = append([]byte{b}, bytes...)
		n >>= 7
		cont = true
	}
	return bytes
}
