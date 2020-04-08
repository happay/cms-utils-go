package util

import (
	"bytes"
	"cms-utils-go/logger"
	"encoding/json"
	"net/http"
)

// APICall will call api specifed in the path with parameters specified..
func APICall(method, path string, requestBody PropertyMap,
	headers map[string]string) (responseCode int, responseBody map[string]string, err error) {

	requestBytes, err := requestBody.Value()
	if err != nil {
		logger.GetLogger().Printf("could not transform the request body %s, err: %s", requestBody, err)
		return
	}

	requestBuffer := bytes.NewBuffer(requestBytes.([]byte))
	req, err := http.NewRequest(method, path, requestBuffer)
	if err != nil {
		logger.GetLogger().Printf("error while making httprequest body for %s, err: %s", requestBody, err)
		return
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.GetLogger().Printf("error while making the request %s, err: %s", requestBody, err)
		return
	}

	defer resp.Body.Close()

	responseCode = resp.StatusCode
	responseBody = make(map[string]string)

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		logger.GetLogger().Printf("error while decoding the response body %s err:%s", resp.Body, err)
		return
	}
	return
}
