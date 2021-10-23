package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"gopkg.in/yaml.v2"
)

type config struct {
	AccessKeyId  string `yaml:"access_key_id"`
	AccessSecret string `yaml:"access_secret"`
	RegionId     string `yaml:"region_id"`
	GetIpUrl     string `yaml:"get_ip_url"`
}

type Ddns struct {
	cfg    config
	client *alidns.Client
}

// GetConfig 获取配置文件
func (d *Ddns) GetConfig(fileName string) error {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &d.cfg)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (d *Ddns) CreateClient() (err error) {
	d.client, err = alidns.NewClientWithAccessKey(d.cfg.RegionId, d.cfg.AccessKeyId, d.cfg.AccessSecret)
	return
}

// 增加解析记录 RR 主机记录 domain 域名
// 参数示例:("www","baidu.com","182.12.1.1")
// https://help.aliyun.com/document_detail/29772.html
func (d Ddns) AddDoma(RR, domain, ip string) (isOk bool, err error) {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Value = ip
	request.Type = "A"
	request.RR = RR
	request.DomainName = domain

	resp, err := d.client.AddDomainRecord(request)
	isOk = resp.IsSuccess()
	return
}

// GetAliIpAndRecordId GetAliRecordIp 获取解析记录的IP和解析记录的recordId
// @1为解析主机记录如：www @2为子域名如: www.baidu.com
func (d Ddns) GetAliIpAndRecordId(RR, subDomain string) (AliIp string, RecordId string, err error) {
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	request.SubDomain = subDomain
	response, err := d.client.DescribeSubDomainRecords(request)
	if err != nil {
		return
	}
	if response.IsSuccess() {
		for _, v := range response.DomainRecords.Record {
			if v.RR == RR {
				AliIp = v.Value
				RecordId = v.RecordId
				return
			}
		}
	}
	err = notFindRecord
	return
}

// UpdateDNS 更新DNS
func (d Ddns) Update(recordId, RR, ip string) (isOk bool, err error) {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordId
	request.RR = RR
	request.Type = "A"
	request.Value = ip
	resp, err := d.client.UpdateDomainRecord(request)
	isOk = resp.IsSuccess()
	return
}

var (
	fileName      string
	recordId      string
	subDomain     string
	wanIp         string
	notFindRecord = errors.New("没有查找到解析记录")
	ddns          = Ddns{}
)

func init() {
	flag.StringVar(&fileName, "file", "config.yml", "输入配置文件的地址,default ./config.yml")
	flag.StringVar(&subDomain, "domain", "", "解析的子域名比如：www.test.com")
	flag.StringVar(&wanIp, "ip", "", "可选参数更新ip,如果不写那么会从配置文件里的url里取得")
	flag.Usage = Usage
	flag.Parse()
}
func Usage() {
	fmt.Println("aliddns version 0.0.1\n 配置文件请自行定义")
	flag.PrintDefaults()
}

// GetWanIP 获取公网ip
func GetWanIP(url string) (wanIp string, err error) {
	resp, err := http.Get(url)
	if err != err {
		return wanIp, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wanIp, err
	}
	wanIp = string(body)
	return wanIp, nil
}

//分割url 示例 @1 www.baidu.com
func splitUrl(subDomain string) (RR, domain string) {
	arrUrl := strings.Split(subDomain, ".")
	RR = arrUrl[0]
	domain = strings.Join(arrUrl[1:], ".")
	return
}
func exit(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:  error: %s", msg, err)
	}
}

func update() {
	RR, domain := splitUrl(subDomain)
	aliIp, recordId, err := ddns.GetAliIpAndRecordId(RR, subDomain)
	aliIp = strings.TrimSpace(aliIp)
	if aliIp == wanIp {
		log.Printf("解析记录相同不用更新\n")
		return
	}
	switch err {
	case notFindRecord:
		log.Printf("%s,正在进行增加\n", err)
		isOk, err := ddns.AddDoma(RR, domain, wanIp)
		if isOk {
			log.Panicf("%s 记录添加成功 ip: %s", subDomain, wanIp)
		}
		if err != nil {
			log.Printf("记录添加失败%s", err)
		}
	case nil:
		isOk, err := ddns.Update(recordId, RR, wanIp)
		if isOk {
			log.Printf("更新成功：%s,oldip: %s newip: %s", subDomain, aliIp, wanIp)
		}
		if err != nil {
			log.Printf("更新失败%s", err)
		}
	default:
		exit(err, "取记录值失败")
	}
}
func main() {
	if subDomain == "" {
		log.Fatalf("参数domain 不能为空")
	}
	err := ddns.GetConfig(fileName)
	exit(err, "配置文件解析失败")

	err = ddns.CreateClient()
	exit(err, "创建alidas client 失败")
	if wanIp == "" {
		wanIp, err = GetWanIP(ddns.cfg.GetIpUrl)
		exit(err, "取外网ip失败")
		wanIp = strings.TrimSpace(wanIp)
	}
	update()
}
