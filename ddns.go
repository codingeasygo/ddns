package ddns

import (
	"fmt"
	"net"
	"time"

	"github.com/codingeasygo/util/debug"
)

type Task struct {
	Name   string
	RR     string
	Domain string
	Delay  time.Duration
	Last   time.Time
}

type DomainSyncer interface {
	Query(rr, domain string) (ipv4, ipv6 net.IP, err error)
	Modify(rr, domain string, ipv4, ipv6 net.IP) (err error)
}

type PublicDiscover interface {
	Discover(rr, domain string) (ipv4, ipv6 net.IP, err error)
}

type DDNS struct {
	Delay    time.Duration
	Syncer   DomainSyncer
	Discover PublicDiscover
	tasks    []*Task
	running  bool
}

func NewDDNS(syncer DomainSyncer, discover PublicDiscover) (ddns *DDNS) {
	ddns = &DDNS{
		Delay:    time.Second,
		Syncer:   syncer,
		Discover: discover,
	}
	return
}

func (d *DDNS) Add(task *Task) {
	d.tasks = append(d.tasks, task)
}

func (d *DDNS) Start() {
	d.running = true
	go d.loopSync()
}

func (d *DDNS) Stop() {
	d.running = false
}

func (d *DDNS) Run() {
	d.running = true
	d.loopSync()
}

func (d *DDNS) loopSync() {
	for d.running {
		d.procSync()
		time.Sleep(d.Delay)
	}
}

func (d *DDNS) procSync() {
	defer func() {
		if perr := recover(); perr != nil {
			ErrorLog("DDNS proc sync fail with %v, callstack is \n%v", perr, debug.CallStatck())
		}
	}()
	for _, task := range d.tasks {
		if time.Since(task.Last) < task.Delay {
			continue
		}
		ipv4New, ipv6New, err := d.Discover.Discover(task.RR, task.Domain)
		if err != nil {
			WarnLog("DDNS(%v) discover public ip by %v.%v fail with %v", task.Name, task.RR, task.Domain, err)
			continue
		}
		ipv4Having, ipv6Having, err := d.Syncer.Query(task.RR, task.Domain)
		if err != nil {
			WarnLog("DDNS(%v) query domain record by %v.%v fail with %v", task.Name, task.RR, task.Domain, err)
		}
		if ipv4New != nil && ipv4Having != nil && ipv4New.String() == ipv4Having.String() &&
			((ipv6New == nil && ipv6Having == nil) ||
				(ipv6New != nil && ipv6Having != nil && ipv6New.String() == ipv6Having.String())) {
			continue
		}
		err = d.Syncer.Modify(task.RR, task.Domain, ipv4New, ipv6New)
		if err != nil {
			WarnLog("DDNS(%v) modify domain record by %v.%v->%v,%v  fail with %v", task.Name, task.RR, task.Domain, ipv4New, ipv6New, err)
			continue
		}
		task.Last = time.Now()
		InfoLog("DDNS(%v) sync domain record by %v.%v->%v,%v success", task.Name, task.RR, task.Domain, ipv4New, ipv6New)
	}
}

type PoolDiscover struct {
	All []PublicDiscover
}

func NewPoolDiscover(all ...PublicDiscover) (pool *PoolDiscover) {
	pool = &PoolDiscover{All: all}
	return
}

func (p *PoolDiscover) Add(discover PublicDiscover) {
	p.All = append(p.All, discover)
}

func (p *PoolDiscover) Discover(rr, domain string) (ipv4, ipv6 net.IP, err error) {
	if len(p.All) < 1 {
		err = fmt.Errorf("not discover")
		return
	}
	for _, discover := range p.All {
		ipv4, ipv6, err = discover.Discover(rr, domain)
		if err == nil && len(ipv4) > 0 {
			break
		}
	}
	return
}
