package slacklog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func reverse(slice []interface{}) {
	halfLen := len(slice) / 2
	for i := 0; i < halfLen; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func DownloadEntitiesToFile(slackToken string, apiMethod string, extraParams map[string]string, jsonKey string, isReverse bool, destJSONFilePath string) error {
	results, err := getPaginatedEntites(slackToken, apiMethod, extraParams, jsonKey, isReverse)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return nil
	}

	destDir := filepath.Dir(destJSONFilePath)
	err = os.MkdirAll(destDir, 0777)
	if err != nil {
		return err
	}

	w, err := os.Create(destJSONFilePath)
	if err != nil {
		return err
	}
	defer w.Close()

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(&results)
	if err != nil {
		return err
	}

	return nil
}

func getPaginatedEntites(slackToken string, apiMethod string, extraParams map[string]string, jsonKey string, isReverse bool) ([]interface{}, error) {
	// Slack API の pagination の仕様: https://api.slack.com/docs/pagination
	var results []interface{}
	var nextCursor string

	pageNum := 1
	for {
		var json map[string]interface{}

		log.Printf("GET %v page %d of %s ... ", extraParams, pageNum, apiMethod)
		params := map[string]string{
			"token":  slackToken,
			"cursor": nextCursor,
		}
		for key, value := range extraParams {
			params[key] = value
		}
		err := httpGetJSON("https://slack.com/api/"+apiMethod, params, &json)
		if err != nil {
			return nil, err
		}

		responseOK := true
		okResponseValue, ok := json["ok"]
		if ok {
			responseOK, ok = okResponseValue.(bool)
			if !ok {
				responseOK = false
			}
		} else {
			responseOK = false
		}
		if !responseOK {
			errorMessage, ok := json["error"].(string)
			if ok {
				return nil, fmt.Errorf("Error response: %s", errorMessage)
			}
			return nil, fmt.Errorf("Error response: %v", json)
		}

		targetItem, ok := json[jsonKey]
		if !ok {
			return nil, fmt.Errorf("Key not found in response:" + jsonKey)
		}
		list, ok := targetItem.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Unknown type of value:" + jsonKey)
		}

		log.Printf("fetched %s count: %d\n", jsonKey, len(list))

		results = append(results, list...)

		nextCursor = getNextCursor(json)
		if nextCursor == "" {
			break
		}
		pageNum += 1
	}

	if isReverse {
		reverse(results)
	}

	return results, nil
}

func getNextCursor(json map[string]interface{}) string {
	metadataValue, ok := json["response_metadata"]
	if !ok {
		return ""
	}
	metadata, ok := metadataValue.(map[string]interface{})
	if !ok {
		return ""
	}
	nextCursorValue, ok := metadata["next_cursor"]
	if !ok {
		return ""
	}
	nextCursor, ok := nextCursorValue.(string)
	if !ok {
		return ""
	}
	return nextCursor
}

func httpGetJSON(rawurl string, params map[string]string, dst interface{}) error {
	url, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	values := url.Query()
	for name, value := range params {
		values.Set(name, value)
	}
	url.RawQuery = values.Encode()

	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("[%s]: %s", resp.Status, url.String())
	}

	err = json.NewDecoder(resp.Body).Decode(dst)
	if err != nil {
		return err
	}

	return nil
}
