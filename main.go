package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/denverdino/aliyungo/dns"
)

const (
	accessKeyId     = "xxxx"
	accessKeySecret = "xxxx"
	domainName      = "tiduyun.com"
	subSomain       = "150"
)

var client *dns.Client

func init() {
	client = dns.NewClient(accessKeyId, accessKeySecret)
}
func catch(fun func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			time.Sleep(5 * time.Second)
			catch(fun)
		}
	}()
	fun()
}
func main() {
	catch(run)
}
func run() {
	describeArgs := dns.DescribeDomainRecordsArgs{
		DomainName: domainName,
	}
	var ip string
	for {
		ip = getLocalIp()
		domain := subSomain + "." + domainName
		nameIp := nslookup(domain)
		log.Printf("adsl ip:%s , nslookup ip:%s,domain:%s \n", ip, nameIp, domain)
		if nameIp == "" || ip == "" || ip == nameIp {
			time.Sleep(10 * time.Second)
			continue
		}
		log.Println("IP不相等，开始更新IP")
		descResponse, err := client.DescribeDomainRecords(&describeArgs)
		checkErr(err)
		for _, descRecord := range descResponse.DomainRecords.Record {
			if ip != descRecord.Value && descRecord.Type == "A" && descRecord.RR == subSomain {
				//fmt.Println(descRecord.RecordId, descRecord.RR, ip)
				update(descRecord.RecordId, descRecord.RR, ip)
				log.Println("更新成功，等待5分钟后再检测！")
				time.Sleep(300 * time.Second)
				break
			} else {
				log.Println("查询到IP与当前IP一致，请等待DNS缓存更新！5分钟后再试")
				time.Sleep(300 * time.Second)
				break
			}
		}
		time.Sleep(10 * time.Second)
	}
}

type IpInfo struct {
	Cip   string
	Cid   string
	Cname string
}

func getLocalIp() string {
	resp, err := http.Get("http://pv.sohu.com/cityjson")
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	body = body[19 : len(body)-1]
	var ipInfo IpInfo
	json.Unmarshal(body, &ipInfo)
	return ipInfo.Cip
}

func update(recordId string, rr string, ip string) {
	updateArgs := dns.UpdateDomainRecordArgs{
		RecordId: recordId,
		RR:       rr,
		Value:    ip,
		Type:     "A",
	}
	log.Println(updateArgs)
	_, err := client.UpdateDomainRecord(&updateArgs)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func nslookup(domain string) string {
	host, err := net.LookupHost(domain)
	if err != nil || len(host) == 0 {
		log.Println("nslookup faile:", err)
		return ""
	}
	return host[0]
}
