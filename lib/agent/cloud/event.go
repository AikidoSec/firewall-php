package cloud

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"main/config"
	"main/log"
	"net/http"
	"net/url"
)

func SendCloudRequest(endpoint string, route string, method string, payload interface{}) ([]byte, error) {
	token := config.GetToken()
	if token == "" {
		return nil, fmt.Errorf("no token set")
	}

	apiEndpoint, err := url.JoinPath(endpoint, route)
	if err != nil {
		return nil, fmt.Errorf("failed to build API endpoint: %v", err)
	}

	var req *http.Request
	if payload != nil {
		var jsonData []byte
		jsonData, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}

		log.Infof("Sending %s request to %s with size %d and content: %s", method, apiEndpoint, len(jsonData), jsonData)

		req, err = http.NewRequest(method, apiEndpoint, bytes.NewBuffer(jsonData))
	} else {
		log.Infof("Sending %s request to %s", method, apiEndpoint)
		req, err = http.NewRequest(method, apiEndpoint, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	// Check if response is Gzip-encoded
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Debugf("Got response: %s", body)
	return body, nil
}
