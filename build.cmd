@REM windows平台编译其他平台程序
@REM SETCGO_ENABLED=0
@REM SETGOOS=darwin
@REM SETGOARCH=amd64
@REM gobuild main.go
SETCGO_ENABLED=0
SETGOOS=linux
SETGOARCH=amd64
go build -o aliyun-ddns main.go