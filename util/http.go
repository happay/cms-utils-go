package util

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"

	"github.com/happay/cms-utils-go/v3/logger"
)

type Options struct {
	publicKey, privateKey []byte
	queryParams           map[string]string
	requestBody           PropertyMap
	header                map[string]string
	timeout               time.Duration
	caCert                []byte
	insecureSkipVerify    bool
}

func MakeHttpRequest(method, path string, opts ...HttpOption) (responseBody *http.Response, err error) {
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

func WithCertificate(publicKey, privateKey, caCert []byte, insecureSkipVerify bool) HttpOption {
	return func(h *Options) {
		h.privateKey = privateKey
		h.publicKey = publicKey
		if len(caCert) != 0 {
			h.caCert = caCert
		}
		h.insecureSkipVerify = insecureSkipVerify
	}
}

func addClientConfig(opt *Options) *http.Client {
	client := &http.Client{}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opt.insecureSkipVerify,
	}
	if len(opt.privateKey) != 0 && len(opt.publicKey) != 0 {
		cert, _ := tls.X509KeyPair(opt.publicKey, opt.privateKey)
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certs (from a provided cert string)
	if len(opt.caCert) != 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(opt.caCert)
		tlsConfig.RootCAs = caCertPool
	}
	// Assign the TLS config to the client's transport if configured
	if len(tlsConfig.Certificates) > 0 || tlsConfig.RootCAs != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}
	if opt.timeout != 0 {
		client.Timeout = opt.timeout
	}
	return client
}

func httpRequest(method, path string, opt *Options) (resp *http.Response, err error) {
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

	logger.GetLoggerV3().Info("[httpRequest] method: , path: ", slog.String("Method", method), slog.String("path", path))
	logger.GetLoggerV3().Info("[httpRequest] request body:", slog.Any("requestBody", opt.requestBody))

	// if query param exits, add
	if len(opt.queryParams) != 0 {
		logger.GetLoggerV3().Info("[httpRequest] query param: ", slog.Any("Query Params", opt.queryParams))
		queryParams := opt.queryParams
		q := req.URL.Query()
		for key, val := range queryParams {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	if len(opt.header) != 0 {
		logger.GetLoggerV3().Info("[httpRequest] Header:  %v\n", slog.Any("Header", opt.header))
		for key, val := range opt.header {
			req.Header.Set(key, val)
		}
	}
	client := addClientConfig(opt)

	resp, err = client.Do(req)
	return
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
