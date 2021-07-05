package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/codingeasygo/ddns"
	"github.com/codingeasygo/ddns/finder"
	"github.com/codingeasygo/util/xprop"
)

type AlidnsSyncer struct {
	having map[string]net.IP
	locker sync.RWMutex
	client *alidns.Client
}

func NewAlidnsSyncer(regionId, accessKeyId, accessKeySecret string) (syncer *AlidnsSyncer, err error) {
	syncer = &AlidnsSyncer{
		having: map[string]net.IP{},
		locker: sync.RWMutex{},
	}
	syncer.client, err = alidns.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	return
}

func (a *AlidnsSyncer) Query(rr, domain string) (ipv4, ipv6 net.IP, err error) {
	a.locker.RLock()
	ipv4 = a.having[fmt.Sprintf("%v.%v", rr, domain)]
	a.locker.RUnlock()
	return
}

func (a *AlidnsSyncer) Modify(rr, domain string, ipv4, ipv6 net.IP) (err error) {
	defer func() {
		if err == nil {
			a.locker.Lock()
			a.having[fmt.Sprintf("%v.%v", rr, domain)] = ipv4
			a.locker.Unlock()
		}
	}()
	checkRequest := alidns.CreateDescribeDomainRecordsRequest()
	checkRequest.Scheme = "https"
	checkRequest.RRKeyWord = rr
	checkRequest.DomainName = domain
	checkResponse, _ := a.client.DescribeDomainRecords(checkRequest)
	if checkResponse == nil || len(checkResponse.DomainRecords.Record) < 1 {
		request := alidns.CreateAddDomainRecordRequest()
		request.Scheme = "https"
		request.Value = ipv4.String()
		request.Type = "A"
		request.RR = rr
		request.DomainName = domain
		_, err = a.client.AddDomainRecord(request)
	} else {
		record := checkResponse.DomainRecords.Record[0]
		if record.Value == ipv4.String() {
			return
		}
		request := alidns.CreateUpdateDomainRecordRequest()
		request.RecordId = record.RecordId
		request.Scheme = "https"
		request.Value = ipv4.String()
		request.Type = "A"
		request.RR = rr
		_, err = a.client.UpdateDomainRecord(request)
	}
	return
}

func main() {
	confPath := "conf/aliddns.properties"
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}
	conf := xprop.NewConfig()
	conf.Load(confPath)
	conf.Print()
	var regionId, accessKeyId, accessKeySecret string
	err := conf.ValidFormat(`
		syncer.alidns/region_id,R|S,L:0;
		syncer.alidns/access_key_id,R|S,L:0;
		syncer.alidns/access_key_secret,R|S,L:0;
	`, &regionId, &accessKeyId, &accessKeySecret)
	if err != nil {
		ddns.ErrorLog("load config fail with %v", err)
		os.Exit(1)
	}
	syncer, err := NewAlidnsSyncer(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		ddns.ErrorLog("parse alidns client fail with %v", err)
		os.Exit(1)
	}
	discover := ddns.NewPoolDiscover()
	runner := ddns.NewDDNS(syncer, discover)
	for _, sec := range conf.Seces {
		if strings.HasPrefix(sec, "task.") {
			name := conf.StrDef(strings.TrimPrefix(sec, "task."), sec+"/name")
			rr := conf.StrDef("", sec+"/rr")
			domain := conf.StrDef("", sec+"/domain")
			delay := conf.Int64Def(30*60, sec+"/delay")
			if len(rr) < 1 || len(domain) < 1 {
				ddns.WarnLog("parse %v task fail with rr/domain not exists", sec)
				continue
			}
			runner.Add(&ddns.Task{
				Name:   name,
				RR:     rr,
				Domain: domain,
				Delay:  time.Duration(delay) * time.Millisecond,
			})
		}
		if strings.HasPrefix(sec, "discover.") {
			name := conf.StrDef(strings.TrimPrefix(sec, "discover."), sec+"/name")
			serverURI := conf.StrDef("", sec+"/server_uri")
			key := conf.StrDef("/ip", sec+"/key")
			if len(serverURI) < 1 {
				ddns.WarnLog("parse %v discover fail with server_uri not exists", sec)
				continue
			}
			f := finder.NewFinderServer(name, serverURI, key)
			f.Key = key
			discover.Add(f)
		}
	}
	runner.Run()
}
