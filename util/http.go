package util

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/happay/cms-utils-go/v2/logger"
)

type Options struct {
	publicKey, privateKey string
	queryParams           map[string]string
	requestBody           PropertyMap
	header                map[string]string
	timeout               time.Duration
}

func MakeHttpRequest(method, path string, opts ...HttpOption) (
	responseCode int, responseBody map[string]interface{}, err error,
) {
	h := &Options{}
	for _, opt := range opts {
		opt(h)
	}
	return httpRequest(method, path, h)
}

type HttpOption func(*Options)

func WithQueryParam(queryParams map[string]string) HttpOption {
	return func(h *Options) {
		h.queryParams = queryParams
	}
}

func WithRequestBody(requestBody PropertyMap) HttpOption {
	return func(h *Options) {
		h.requestBody = requestBody
	}
}

func WithHeader(header map[string]string) HttpOption {
	return func(h *Options) {
		h.header = header
	}
}

func WithTimeoutInSec(timeout int64) HttpOption {
	return func(h *Options) {
		h.timeout = time.Duration(timeout) * time.Second
	}
}

func WithCertificate(publicKey, privateKey string) HttpOption {
	return func(h *Options) {
		h.privateKey = privateKey
		h.publicKey = publicKey
	}
}

func addClientConfig(opt *Options) *http.Client {
	client := &http.Client{}
	if strings.TrimSpace(opt.privateKey) != "" && strings.TrimSpace(opt.publicKey) != "" {
		cert, _ := tls.LoadX509KeyPair(opt.publicKey, opt.privateKey)
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}
	if opt.timeout != 0 {
		client.Timeout = opt.timeout
	}
	return client
}

func httpRequest(method, path string, opt *Options) (responseCode int, responseBody map[string]interface{}, err error) {
	requestBody := opt.requestBody
	requestBytes, err := requestBody.Value()
	if err != nil {
		return
	}
	requestBuffer := bytes.NewBuffer(requestBytes.([]byte))
	req, err := http.NewRequest(method, path, requestBuffer)
	if err != nil {
		return
	}

	logger.GetLoggerV3().Info("[httpRequest] method:  %s, path: %s\n", method, path)
	logger.GetLoggerV3().Info("[httpRequest] request body:  %v\n", opt.requestBody)

	// if query param exits, add
	if len(opt.queryParams) != 0 {
		logger.GetLoggerV3().Info("[httpRequest] query param:  %v\n", opt.queryParams)
		queryParams := opt.queryParams
		q := req.URL.Query()
		for key, val := range queryParams {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	if len(opt.header) != 0 {
		logger.GetLoggerV3().Info("[httpRequest] Header:  %v\n", opt.queryParams)
		for key, val := range opt.header {
			req.Header.Set(key, val)
		}
	}
	client := addClientConfig(opt)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	responseCode = resp.StatusCode
	responseBody = make(map[string]interface{})

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	logger.GetLoggerV3().Info("[httpRequest] response body:  %v\n", responseBody)
	return
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
