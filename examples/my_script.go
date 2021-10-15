package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func postResultToRuntime(respURL *string, profileId *string) {
	var jsonStr = []byte(fmt.Sprintf(`{"text":"Customer data for %s: <add data here>"}`, *profileId))
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
	profileIdPtr := flag.String("profile_id", "none...", "Profile ID of user")
	flag.Parse()

	// Do stuff

	postResultToRuntime(&url, profileIdPtr)
}
