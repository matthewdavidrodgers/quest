package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/matthewdavidrodgers/quest/internal/config"
	"github.com/matthewdavidrodgers/quest/internal/editor"
	"github.com/matthewdavidrodgers/quest/internal/http"
)

var openEditor bool
var formatOutput bool
var method string
var verbose bool
var redo bool

var url string

func formatJSON(data string) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(data), "", "  ")
	if err != nil {
		return "", errors.New("Error: response data is not valid JSON; use -f=false to display raw")
	}
	return out.String(), nil
}

func init() {
	const (
		openEditorCmd     = "e"
		openEditorDefault = false
		openEditorUsage   = "Open an editor to fill in request details before sending"

		formatOutputCmd     = "f"
		formatOutputDefault = true
		formatOutputUsage   = "Format response body as json"

		methodCmd     = "m"
		methodDefault = "GET"
		methodUsage   = "HTTP request method. GET, PUT, POST, PATCH, DELETE allowed"

                verboseCmd     = "v"
                verboseDefault = false
                verboseUsage   = "Output verbose information about the request and response"

                redoCmd     = "r"
                redoDefault = false
                redoUsage   = "Use the same details as the previous request"
	)

	flag.BoolVar(&openEditor, openEditorCmd, openEditorDefault, openEditorUsage)
	flag.BoolVar(&formatOutput, formatOutputCmd, formatOutputDefault, formatOutputUsage)
	flag.StringVar(&method, methodCmd, methodDefault, methodUsage)
        flag.BoolVar(&verbose, verboseCmd, verboseDefault, verboseUsage)
        flag.BoolVar(&redo, redoCmd, redoDefault, redoUsage)
}

func main() {
	configFile, err := config.OpenConfigFile(".questconfig")
	defer configFile.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) == 4 && os.Args[1] == "cookie" && os.Args[2] == "add" {
		value := os.Args[3]
		if err := configFile.AddCookieToStore(value); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	} else if len(os.Args) == 3 && os.Args[1] == "cookie" && os.Args[2] == "wipe" {
		if err := configFile.RemoveCookie(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	flag.Parse()
	posArgCount := flag.NArg()

	if posArgCount > 1 {
		flag.Usage()
		os.Exit(1)
	} else if posArgCount == 1 {
		url = flag.Arg(0)
	}

	req := http.NewRequestDetails()
	req.Method = method
	req.Url = url

        if redo {
                lastReqEncoded, err := configFile.GetValue("LAST_REQUEST")
                if err != nil {
                        fmt.Println(err)
                        os.Exit(1)
                }
                if lastReqEncoded == "" {
                        fmt.Println("No last request found, cannot use -r")
                        os.Exit(1)
                }
                lastReq := editor.EncodeableRequestDetails{}
                err = json.Unmarshal([]byte(lastReqEncoded), &lastReq)
                if err != nil {
                        fmt.Println(err)
                        os.Exit(1)
                }
                req.Method = lastReq.Method
                req.Url = lastReq.Url
                req.Headers = lastReq.Headers
                if lastReq.Body != nil {
                        req.Body = string(*lastReq.Body)
                }
        }

	if openEditor {
		err = editor.OpenEditorAndParseFromSavedContent(&req)
		if err != nil {
			fmt.Println("Error with editor: ", err)
			os.Exit(1)
		}
	}

	cookie, err := configFile.GetCookie()
	if err != nil {
		panic(err)
	}
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}

        var rawBody json.RawMessage
        if req.Body != "" {
                rawBody = []byte(req.Body)
        }
        encoded, err := json.Marshal(editor.EncodeableRequestDetails{
                Method: req.Method,
                Url: req.Url,
                Headers: req.Headers,
                Body: &rawBody,
        })
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
        configFile.RemoveValue("LAST_REQUEST")
        _, err = configFile.AppendValue("LAST_REQUEST", string(encoded))
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }

        output := ""
        if verbose {
                reqDetails := http.PrintRequest(req)
                output = output + reqDetails
        }

	resp, err := http.MakeRequest(req)
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }

        content, err := http.ReadResponse(resp, verbose, formatOutput)
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
        output = output + content

        fmt.Println(output)
}
