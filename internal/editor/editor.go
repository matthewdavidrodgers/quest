package editor

import (
        "io"

        "encoding/json"
        "os"
        "os/exec"

        "github.com/matthewdavidrodgers/quest/internal/http"
)

type EncodeableRequestDetails struct {
        Method  string            `json:"method"`
        Url     string            `json:"name"`
        Headers map[string]string `json:"headers"`
        Body    *json.RawMessage  `json:"body"`
}

func OpenEditorAndParseFromSavedContent(req *http.RequestDetails) error {
        requestFile, err := os.Create("quest_editor_buf.json")
        requestFileName := requestFile.Name()
        if err != nil {
                return err
        }

        var emptyBody json.RawMessage = []byte("{}")
        encodeableData := EncodeableRequestDetails{
                Method: req.Method,
                Url: req.Url,
                Headers: req.Headers,
                Body: &emptyBody,
        }
        encodedReq, err := json.MarshalIndent(encodeableData, "", "  ")
        if err != nil {
                return err
        }

        _, err = requestFile.Write(encodedReq)
        if err != nil {
                return err
        }
        requestFile.Close()

        editorProcess := exec.Command("vim", requestFileName)

        editorProcess.Stdout = os.Stdout
        editorProcess.Stdin = os.Stdin
        editorProcess.Stderr = os.Stderr

        editorProcess.Run()
        if err != nil {
                return err
        }

        requestFile, err = os.Open(requestFileName)
        defer requestFile.Close()
        defer os.Remove(requestFileName)
        if err != nil {
                return err
        }

        contents, err := io.ReadAll(requestFile)
        if err != nil {
                return err
        }

        err = json.Unmarshal(contents, &encodeableData)
        if err != nil {
                return err
        }
        req.Method = encodeableData.Method
        req.Url = encodeableData.Url
        req.Headers = encodeableData.Headers
        req.Body = string(*encodeableData.Body)

        return nil
}
