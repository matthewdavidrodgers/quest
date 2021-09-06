package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RequestDetails struct {
	Method  string
	Url     string
	Headers map[string]string
	Body    string
}

type ResponseError struct {
	StatusCode    int
	StatusMessage string
	Body          string
}

func (e *ResponseError) Error() string {
	err := fmt.Sprintf("Failed with %s", e.StatusMessage)
	if e.Body != "" {
		err += "\n" + e.Body
	}
	return err
}

func NewRequestDetails() RequestDetails {
	return RequestDetails{
		Method:  "",
		Url:     "",
		Headers: make(map[string]string),
		Body:    "",
	}
}

var client = &http.Client{}

func MakeRequest(reqDetails RequestDetails) (string, error) {
	bodyReader := strings.NewReader(reqDetails.Body)
	req, err := http.NewRequest(reqDetails.Method, reqDetails.Url, bodyReader)
	if err != nil {
		return "", err
	}
	for name, value := range reqDetails.Headers {
		req.Header.Add(name, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", &ResponseError{
			StatusCode:    resp.StatusCode,
			StatusMessage: resp.Status,
			Body:          string(data),
		}
	}
	if err != nil {
		return "", err
	}

	return string(data), err
}
