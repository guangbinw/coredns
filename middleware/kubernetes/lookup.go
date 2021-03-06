package kubernetes

import (
	"fmt"
	"net"

	"github.com/coredns/coredns/middleware/etcd/msg"
	"github.com/coredns/coredns/middleware/pkg/dnsutil"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func (k Kubernetes) records(state request.Request, exact bool) ([]msg.Service, error) {
	r, err := k.parseRequest(state.Name(), state.QType())
	if err != nil {
		return nil, err
	}
	services, err := k.Records(r)
	if err != nil {
		return nil, err
	}
	services = msg.Group(services)
	return services, nil
}

// PTR Record returns PTR records from kubernetes.
func (k Kubernetes) PTR(zone string, state request.Request) ([]dns.RR, error) {
	reverseIP := dnsutil.ExtractAddressFromReverse(state.Name())
	if reverseIP == "" {
		return nil, fmt.Errorf("does not support reverse lookup for %s", state.QName())
	}

	records := make([]dns.RR, 1)
	services, err := k.records(state, false)
	if err != nil {
		return nil, err
	}

	for _, serv := range services {
		ip := net.ParseIP(serv.Host)
		if reverseIP != serv.Host {
			continue
		}
		switch {
		case ip.To4() != nil:
			records = append(records, serv.NewPTR(state.QName(), ip.To4().String()))
			break
		case ip.To4() == nil:
			// nodata?
		}
	}
	return records, nil
}
