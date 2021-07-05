package finder

import (
	"fmt"
	"net"

	"github.com/codingeasygo/ddns"
	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
)

type FinderServer struct {
	Name      string
	ServerURI string
	Key       string
}

func NewFinderServer(name, serverURI, key string) (server *FinderServer) {
	server = &FinderServer{
		Name:      name,
		ServerURI: serverURI,
		Key:       key,
	}
	return
}

func (f *FinderServer) Discover(rr, domain string) (ipv4, ipv6 net.IP, err error) {
	res, err := xhttp.GetMap("%v", f.ServerURI)
	if err != nil {
		ddns.InfoLog("FinderServer(%v) request %v fail with %v", f.Name, f.ServerURI, err)
		return
	}
	ipv4 = net.ParseIP(res.Str(f.Key))
	if len(ipv4) < 1 {
		ddns.WarnLog("WebFinder(%v) parse ipv4 matcher fail with %v by %v ", err, converter.JSON(res))
		err = fmt.Errorf("parse fail")
	}
	return
}
