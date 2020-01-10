package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	i IP
	s string
}

func TestRecordSuccEncode(t *testing.T) {
	s := IpSerializer{}

	succCases := getSuccCases()

	for _, c := range succCases {
		got, err := s.Encode(c.s)
		assert.NoError(t, err)
		assert.Equal(t, c.i, got)
	}

}

func TestRecordErrEncode(t *testing.T) {
	s := IpSerializer{}

	errCases := []string{"", "-1.1.1.1", "1.1.1.1.1.1", "1.1.1", "0.0.0.256", "asdf"}

	for _, i := range errCases {
		got, err := s.Encode(i)
		assert.Equal(t, got, IP(0))
		assert.NotNil(t, err)
	}
}

func TestRecordSuccDecode(t *testing.T) {
	s := IpSerializer{}

	succCases := getSuccCases()

	for _, c := range succCases {
		got := s.Decode(c.i)
		assert.Equal(t, c.s, got)
	}
}

func getSuccCases() []testCase {
	return []testCase{
		{0, "0.0.0.0"},
		{16843009, "1.1.1.1"},
		{4294967295, "255.255.255.255"},
	}
}
