package main

import (
	"bytes"
	"net/http"
	"os"
)

func postResultToRuntime(respURL *string) {
	var jsonStr = []byte(`{"blocks":{"type":"section","text":{"type":"mrkdwn","text":"Hey there"}}}`)
	req, _ := http.NewRequest("POST", *respURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	url := os.Getenv("RUNTIME_RESPONSE_URL")

	// Do stuff

	postResultToRuntime(&url)
}
