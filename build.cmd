@REM windows平台编译其他平台程序
@REM SETCGO_ENABLED=0
@REM SETGOOS=darwin
@REM SETGOARCH=amd64
@REM gobuild main.go
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o aliyun-ddns main.go