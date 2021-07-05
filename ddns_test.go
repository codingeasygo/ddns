package ddns

import (
	"fmt"
	"net"
	"testing"
	"time"
)

type testImpl struct {
	ipv4        net.IP
	ipv6        net.IP
	queryErr    bool
	modifyErr   bool
	discoverErr bool
}

func (t *testImpl) Query(rr, domain string) (ipv4, ipv6 net.IP, err error) {
	if t.queryErr {
		err = fmt.Errorf("error")
		return
	}
	ipv4, ipv6 = t.ipv4, t.ipv6
	return
}

func (t *testImpl) Modify(rr, domain string, ipv4, ipv6 net.IP) (err error) {
	if t.modifyErr {
		err = fmt.Errorf("error")
		return
	}
	t.ipv4, t.ipv6 = ipv4, ipv6
	return
}

func (t *testImpl) Discover(rr, domain string) (ipv4, ipv6 net.IP, err error) {
	if t.discoverErr {
		err = fmt.Errorf("error")
		return
	}
	ipv4 = net.ParseIP("127.0.0.1")
	return
}

func TestDDNS(t *testing.T) {
	impl := &testImpl{}
	ddns := NewDDNS(impl, impl)
	ddns.Delay = 10 * time.Millisecond
	ddns.Start()
	ddns.Add(&Task{
		RR:     "x",
		Domain: "test.com",
		Delay:  20 * time.Millisecond,
	})
	time.Sleep(100 * time.Millisecond)
	ddns.Stop()
	time.Sleep(100 * time.Millisecond)
	//
	go ddns.Run()
	time.Sleep(100 * time.Millisecond)
	ddns.Stop()
	time.Sleep(100 * time.Millisecond)
	//
	//test error
	impl.discoverErr = true
	impl.queryErr = false
	impl.modifyErr = false
	ddns.procSync()
	impl.discoverErr = false
	impl.queryErr = true
	impl.modifyErr = true
	ddns.procSync()
	//
	ddns.Add(nil)
	ddns.procSync()
}

func TestPoolDiscover(t *testing.T) {
	pool := NewPoolDiscover()
	_, _, err := pool.Discover("", "")
	if err == nil {
		t.Error(err)
		return
	}
	pool.Add(&testImpl{})
	_, _, err = pool.Discover("", "")
	if err != nil {
		t.Error(err)
		return
	}
}
