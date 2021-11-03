package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func postResultToRuntime(token *string, profileId *string) {
	var jsonStr = []byte(fmt.Sprintf(`{"text":"Customer data for %s: <add data here>"}`, *profileId))
	url := "https://runtime-3aefrytd7wlprd39j.au.ngrok.io/api/response"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	token := os.Getenv("RESPONSE_TOKEN")
	profileId := flag.String("profile_id", "none...", "Profile ID of user")
	flag.Parse()

	// Get customer data
	// ...

	// Post the result to runtime
	postResultToRuntime(&token, profileId)
}
