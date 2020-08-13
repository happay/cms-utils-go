package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// APICall will call api specifed in the path with parameters specified..
func APICall(method, path string, requestBody PropertyMap,
	headers map[string]string) (responseCode int, responseBody PropertyMap, err error) {

	requestBytes, err := requestBody.Value()
	if err != nil {
		return
	}

	requestBuffer := bytes.NewBuffer(requestBytes.([]byte))
	req, err := http.NewRequest(method, path, requestBuffer)
	if err != nil {
		return
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	responseCode = resp.StatusCode

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return
	}
	return
}
