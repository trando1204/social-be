package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	httpClient *http.Client
	cancelFunc context.CancelFunc
	context    context.Context
}

type ReqConfig struct {
	Payload  interface{}
	Method   string
	HttpUrl  string
	Header   map[string]string
	FormData url.Values
}

const defaultHttpClientTimeout = 30 * time.Second

// newClient configures and returns a new client
func newClient() (c *HttpClient) {
	// Initialize context use to cancel all pending requests when shutdown request is made.
	ctx, cancel := context.WithCancel(context.Background())

	return &HttpClient{
		context:    ctx,
		cancelFunc: cancel,
		httpClient: &http.Client{
			Timeout:   defaultHttpClientTimeout,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}
}

func (c *HttpClient) getRequestBody(method string, body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	if method == http.MethodPost {
		if requestBody, ok := body.([]byte); ok {
			return requestBody, nil
		}
	} else if method == http.MethodGet {
		if requestBody, ok := body.(map[string]string); ok {
			params := url.Values{}
			for key, val := range requestBody {
				params.Add(key, val)
			}
			return []byte(params.Encode()), nil
		}
	}

	return nil, errors.New("invalid request body")
}

// query prepares and process HTTP request to backend resources.
func (c *HttpClient) query(reqConfig *ReqConfig) (resp *http.Response, err error) {
	// package the request body for POST and PUT requests
	var requestBody []byte
	if reqConfig.Payload != nil {
		requestBody, err = c.getRequestBody(reqConfig.Method, reqConfig.Payload)
		if err != nil {
			return nil, err
		}
	}

	var body io.Reader
	if requestBody != nil {
		if reqConfig.Method == http.MethodGet {
			reqConfig.HttpUrl += "?" + string(requestBody)
		} else {
			body = bytes.NewReader(requestBody)
		}
	}

	// Create http request
	req, err := http.NewRequestWithContext(c.context, reqConfig.Method, reqConfig.HttpUrl, body)
	if err != nil {
		return nil, fmt.Errorf("error creating http request: %v", err)
	}

	if req == nil {
		return nil, errors.New("error: nil request")
	}

	if reqConfig.Method == http.MethodPost || reqConfig.Method == http.MethodPut {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	} else {
		req.Header.Add("Accept", "application/json")
	}

	for k, v := range reqConfig.Header {
		req.Header.Add(k, v)
	}

	// Send request
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("error: status: %v", resp.Status)
	}

	return resp, nil
}

// HttpRequest queries the API provided in the ReqConfig object and converts
// the returned json(Byte data) into an respObj interface.
func HttpRequest(reqConfig *ReqConfig, respObj interface{}) error {
	client := newClient()

	httpResp, err := client.query(reqConfig)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(httpResp.Body)
	if err := dec.Decode(respObj); err != nil {
		return err
	}

	httpResp.Body.Close()
	return nil
}

type Error struct {
	Code    int    `json:"code" example:"27"`
	Message string `json:"message" example:"Error message"`
}

func GetHttpPost(reqConfig *ReqConfig, respObj interface{}) error {
	resp, err := http.PostForm(reqConfig.HttpUrl, reqConfig.FormData)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&respObj)
		return nil
	} else {
		var e Error
		json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("Error Status: %d, Msg: %s", e.Code, e.Message)
	}
}

func HttpPost(httpUrl string, formData url.Values, respObj interface{}) error {
	req := &ReqConfig{
		Method:   http.MethodPost,
		HttpUrl:  httpUrl,
		FormData: formData,
	}
	if err := GetHttpPost(req, &respObj); err != nil {
		return err
	}
	return nil
}

func HttpFullPost(httpUrl string, body io.ReadCloser, headers map[string]string, respObj interface{}) error {
	req, err := http.NewRequest("POST", httpUrl, body)
	if err != nil {
		return err
	}
	for key := range headers {
		req.Header.Set(key, headers[key])
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(respObj)
	if err != nil {
		return err
	}
	return nil
}
