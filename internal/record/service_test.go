package record

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type serviceTestCase struct {
	a   []IP
	b   []IP
	n   int
	res bool
}

func TestService_HasNCommons(t *testing.T) {
	cases := []serviceTestCase{
		serviceTestCase{[]IP{}, []IP{}, 1, false},
		serviceTestCase{[]IP{}, []IP{1}, 1, false},
		serviceTestCase{[]IP{1}, []IP{1}, 1, true},
		serviceTestCase{[]IP{1}, []IP{2}, 1, false},
		serviceTestCase{[]IP{1, 2}, []IP{3, 5}, 1, false},

		serviceTestCase{[]IP{1}, []IP{1}, 2, false},
		serviceTestCase{[]IP{1}, []IP{2, 3, 4, 5}, 2, false},
		serviceTestCase{[]IP{1, 2}, []IP{2, 3}, 2, false},
		serviceTestCase{[]IP{1, 2, 3, 4}, []IP{4, 5, 6, 7}, 2, false},
		serviceTestCase{[]IP{1, 2, 3, 4}, []IP{3, 4, 5, 6, 7}, 2, true},
		serviceTestCase{[]IP{1, 2, 3}, []IP{2, 3, 4, 5, 6, 7}, 2, true},
	}
	s := &service{}
	for _, c := range cases {
		assert.Equal(t, c.res, s.hasNCommons(c.a, c.b, c.n), fmt.Sprintf("a: %v, b: %v, n: %d", c.a, c.b, c.n))
	}
}
