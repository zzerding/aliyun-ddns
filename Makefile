linux-amd64-ddns: mac-ddns
	GOOS=linux GOARCH=amd64  go build  -o linux-amd64-ddns .
mac-ddns:
	go build -o mac-ddns .
