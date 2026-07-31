package cloud

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	. "main/aikido_types"
	"main/config"
	"main/constants"
	"main/log"
	"main/utils"
	"net/http"
	"net/url"
	"time"
)

func SendCloudRequest(server *ServerData, endpoint string, route string, method string, payload interface{}) ([]byte, error) {
	token := config.GetToken(server)
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

		log.Infof(server.Logger, "[%s] Sending %s request to %s with size %d and content: %s", utils.AnonymizeToken(token), method, apiEndpoint, len(jsonData), jsonData)

		req, err = http.NewRequest(method, apiEndpoint, bytes.NewBuffer(jsonData))
	} else {
		log.Infof(server.Logger, "[%s] Sending %s request to %s", utils.AnonymizeToken(token), method, apiEndpoint)
		req, err = http.NewRequest(method, apiEndpoint, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	// Bounded, so a cloud that accepts the connection but never answers cannot block the calling routine forever
	client := &http.Client{Timeout: constants.CloudRequestTimeoutInMs * time.Millisecond}
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

	log.Debugf(server.Logger, "Got response: %s", body)
	return body, nil
}
