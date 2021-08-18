package main

import (
	"bytes"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
)

type config struct {
	AccessKeyId  string `yaml:"access_key_id"`
	AccessSecret string `yaml:"access_secret"`
	DnsDomain    string `yaml:"dns_domain"`
	AliyunDomain string `yaml:"aliyun_domain"`
	CurlDomain   string `yaml:"curl_domain"`
}

// GetConfig 获取配置文件
func GetConfig() *config {
	yamlFile, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	c := &config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return c
}

// GetWanIP 获取公网ip
func GetWanIP() (wanIp string) {
	var result bytes.Buffer
	cmd := exec.Command("curl", GetConfig().CurlDomain)
	cmd.Stdout = &result
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	wanIp = result.String()[:len(result.String())-1]
	return wanIp
}

// CreateAliDns 创建新的阿里云dns解析,返回记录id
func CreateAliDns() {
	var request *alidns.AddDomainRecordRequest
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", GetConfig().AccessKeyId, GetConfig().AccessSecret)

	request = alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Value = GetWanIP()
	request.Type = "A"
	request.RR = GetConfig().DnsDomain
	request.DomainName = GetConfig().AliyunDomain

	if client != nil {
		_, err = client.AddDomainRecord(request)
	}
	if err != nil {
		log.Println(err)
		return
	}
}

// GetAliIpAndRecordId GetAliRecordIp 获取解析记录的IP和解析记录的recordId
func GetAliIpAndRecordId() (AliIp, RecordId string) {
	var request *alidns.DescribeSubDomainRecordsRequest
	var response *alidns.DescribeSubDomainRecordsResponse
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", GetConfig().AccessKeyId, GetConfig().AccessSecret)

	request = alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	domain := GetConfig().DnsDomain + "." + GetConfig().AliyunDomain
	request.SubDomain = domain
	if client != nil {
		response, err = client.DescribeSubDomainRecords(request)
	}
	if err != nil {
		log.Println(err)
		return "", ""
	}
	if response != nil {
		if response.IsSuccess() {
			for _, v := range response.DomainRecords.Record {
				if v.RR == GetConfig().DnsDomain {
					return v.Value, v.RecordId
				}
			}
		}
	}
	return "", ""
}

// UpdateDNS 更新DNS
func UpdateDNS(recordId string) error {
	var request *alidns.UpdateDomainRecordRequest
	var response *alidns.UpdateDomainRecordResponse
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", GetConfig().AccessKeyId, GetConfig().AccessSecret)

	request = alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = recordId
	request.RR = GetConfig().DnsDomain
	request.Type = "A"
	request.Value = GetWanIP()
	request.Lang = "en"
	request.UserClientIp = GetWanIP()
	request.TTL = "600"
	request.Priority = "1"
	request.Line = "default"

	if client != nil {
		response, err = client.UpdateDomainRecord(request)
		log.Println(response)
	}

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SetDns() {
	AliIp, RecordId := GetAliIpAndRecordId()
	wanIp := GetWanIP()
	if AliIp != wanIp {
		err := UpdateDNS(RecordId)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	//获取记录之前先判断有没有记录
	AliIp, RecordId := GetAliIpAndRecordId()
	if AliIp == "" && RecordId == "" {
		CreateAliDns()
	}
	for {
		go SetDns()
		time.Sleep(time.Hour * 24)
	}
}
