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
	)

	flag.BoolVar(&openEditor, openEditorCmd, openEditorDefault, openEditorUsage)

	flag.BoolVar(&formatOutput, formatOutputCmd, formatOutputDefault, formatOutputUsage)

	flag.StringVar(&method, methodCmd, methodDefault, methodUsage)
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

	resp, err := http.MakeRequest(req)
	if err != nil {
		if responseErr, ok := err.(*http.ResponseError); ok {
			errorBody := responseErr.Body
			formattedErrorBody, formatErr := formatJSON(errorBody)
			if errorBody != "" && formatErr == nil {
				errorBody = formattedErrorBody
			}
			fmt.Println("Failed with", responseErr.StatusMessage)
			if errorBody != "" {
				fmt.Println(errorBody)
			}
			os.Exit(1)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if formatOutput {
		formatted, err := formatJSON(resp)
		if err != nil {
			fmt.Println(resp) // print unformatted; ignore silently
		} else {
			fmt.Println(formatted)
		}
	} else {
		fmt.Println(resp)
	}
}
