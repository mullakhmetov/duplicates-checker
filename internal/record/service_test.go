package record

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type serviceTestCase struct {
	a   []net.IP
	b   []net.IP
	n   int
	res bool
}

var ip1 = net.ParseIP("1.1.1.1")
var ip2 = net.ParseIP("2.2.2.2")
var ip3 = net.ParseIP("3.3.3.3")
var ip4 = net.ParseIP("4.4.4.4")
var ip5 = net.ParseIP("5.5.5.5")
var ip6 = net.ParseIP("6.6.6.6")
var ip7 = net.ParseIP("7.7.7.7")

func TestService_hasNCommons(t *testing.T) {
	cases := []serviceTestCase{
		serviceTestCase{[]net.IP{}, []net.IP{}, 1, false},
		serviceTestCase{[]net.IP{}, []net.IP{ip1}, 1, false},
		serviceTestCase{[]net.IP{ip1}, []net.IP{ip1}, 1, true},
		serviceTestCase{[]net.IP{ip1}, []net.IP{ip2}, 1, false},
		serviceTestCase{[]net.IP{ip1, ip2}, []net.IP{ip3, ip5}, 1, false},

		serviceTestCase{[]net.IP{ip1}, []net.IP{ip1}, 2, false},
		serviceTestCase{[]net.IP{ip1}, []net.IP{ip2, ip3, ip4, ip5}, 2, false},
		serviceTestCase{[]net.IP{ip1, ip2}, []net.IP{ip2, ip3}, 2, false},
		serviceTestCase{[]net.IP{ip1, ip2, ip3, ip4}, []net.IP{ip4, ip5, ip6, ip7}, 2, false},
		serviceTestCase{[]net.IP{ip1, ip2, ip3, ip4}, []net.IP{ip3, ip4, ip5, ip6, ip7}, 2, true},
		serviceTestCase{[]net.IP{ip1, ip2, ip3}, []net.IP{ip2, ip3, ip4, ip5, ip6, ip7}, 2, true},
	}
	s := &service{}
	for _, c := range cases {
		assert.Equal(t, c.res, s.hasNCommons(c.a, c.b, c.n), fmt.Sprintf("a: %v, b: %v, n: %d", c.a, c.b, c.n))
	}
}
