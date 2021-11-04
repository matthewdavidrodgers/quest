package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
        "bytes"
        "errors"
        "encoding/json"
)

func formatJSON(data string) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(data), "", "  ")
	if err != nil {
		return "", errors.New("Error: response data is not valid JSON; use -f=false to display raw")
	}
	return out.String(), nil
}

type RequestDetails struct {
	Method  string
	Url     string
	Headers map[string]string
	Body    string
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

func PrintRequest(reqDetails RequestDetails) string {
        var b strings.Builder

        b.WriteString("#### REQUEST INFO ####\n")
        fmt.Fprintf(&b, "%s %s\n", reqDetails.Method, reqDetails.Url)
        for name, value := range reqDetails.Headers {
                fmt.Fprintf(&b, "%s: %s\n", name, value)
        }
        if reqDetails.Body != "" {
                body := reqDetails.Body
                if formattedBody, err := formatJSON(body); err != nil {
                        body = formattedBody
                }
                b.WriteString(body)
        }

        return b.String()
}

func MakeRequest(reqDetails RequestDetails) (*http.Response, error) {
	bodyReader := strings.NewReader(reqDetails.Body)
	req, err := http.NewRequest(reqDetails.Method, reqDetails.Url, bodyReader)
	if err != nil {
		return nil, err
	}
	for name, value := range reqDetails.Headers {
		req.Header.Add(name, value)
	}

	resp, err := client.Do(req)
        return resp, err
}

func ReadResponse(resp *http.Response, verbose bool, formatOutput bool) (string, error) {
        var b strings.Builder

        if verbose {
                b.WriteString("#### RESPONSE INFO ####\n")
                fmt.Fprintf(&b, "Status: %s\n", resp.Status)
                for name, values := range resp.Header {
                        fmt.Fprintf(&b, "%s: ", name)
                        for i, value := range values {
                                if i != 0 {
                                        fmt.Fprintf(&b, ", %s", value)
                                } else {
                                        b.WriteString(value)
                                }
                        }
                        b.WriteRune('\n')
                }
        }

        data, err := io.ReadAll(resp.Body)
        if err != nil {
                return "", err
        }
        if resp.StatusCode < 200 || resp.StatusCode > 299 {
                fmt.Fprintf(&b, "Failed with: %s\n", resp.Status)
                if formatOutput {
                        errorBody := string(data)
                        formattedErrorBody, formatErr := formatJSON(errorBody)
                        if errorBody != "" && formatErr == nil {
                                errorBody = formattedErrorBody
                        }
                        b.WriteString(errorBody)
                }
        } else {
                body := string(data)
                if formatOutput {
                        formatted, err := formatJSON(body)
                        if err == nil {
                                body = formatted
                        }
                }
                b.WriteString(body)
        }

        return b.String(), nil
}
