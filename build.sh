CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
gobuild main.go

CGO_ENABLED=0
GOOS=windows
GOARCH=amd64
gobuild main.go

CGO_ENABLED=0
GOOS=darwin
GOARCH=amd64
gobuild main.go

go build -o aliyun-ddns main.go
nohup ./aliyun-ddns >aliyun-ddns.log 2>&1 &
