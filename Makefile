build: 
	GOOS=linux GOARCH=386 go build cmd/agent/main.go && mv main ./dist/clickport-agent