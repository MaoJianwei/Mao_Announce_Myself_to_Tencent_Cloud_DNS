package main

import (
	"MaoAnnounceMyself/util"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	secretId := os.Getenv("TENCENT_CLOUD_ID") // flag.String("i", "", "secretId")
	secretKey := os.Getenv("TENCENT_CLOUD_KEY") // flag.String("k", "", "secretKey")
	//flag.Parse()
	credential := common.NewCredential(secretId, secretKey)

	for {
		time.Sleep(1 * time.Second)

		client, err := dnspod.NewClient(credential, regions.Beijing, profile.NewClientProfile())
		if err != nil {
			log.Printf("Fail to create client, %s", err)
			continue
		}

		for {
			var recordId *uint64
			req := dnspod.NewDescribeRecordListRequest()
			req.Domain = common.StringPtr("maojianwei.com")
			resp, err := client.DescribeRecordList(req) // API Limit: 100QPS
			if err != nil {
				log.Printf("Fail to get record list, %s", err.Error())
				break
			}
			for _, item := range resp.Response.RecordList {
				if *item.Name == "server" && *item.Type == "AAAA" {
					log.Printf("Get it, %d", *item.RecordId)
					recordId = item.RecordId
					break
				}
			}

			var v6Ip net.IP = nil
			ips, err := util.GetUnicastIp()
			if err != nil {
				break
			}
			for _, v6 := range ips {
				if v6IpTmp := net.ParseIP(v6); v6IpTmp != nil {
					if util.JudgeIPv6(&v6IpTmp) {
						log.Printf("Find IPv6 Unicast, %s", v6IpTmp.String())
						v6Ip = v6IpTmp
						break
					}
				}
			}

			if v6Ip != nil {
				request := dnspod.NewModifyRecordRequest()
				request.Domain = common.StringPtr("maojianwei.com")
				request.SubDomain = common.StringPtr("server")
				request.RecordType = common.StringPtr("AAAA")
				request.RecordLine = common.StringPtr("默认")
				request.Value = common.StringPtr(v6Ip.String())
				request.RecordId = recordId

				response, err := client.ModifyRecord(request) // API Limit: 500QPS
				if err != nil {
					log.Printf("Fail to modify record, %s", err.Error())
					break
				} else {
					log.Printf("Modify record ok, %s", response.ToJsonString())
				}
			}
		}
	}
}
