go build -o aliyun-ddns main.go
nohup ./aliyun-ddns >aliyun-ddns.log 2>&1 &
