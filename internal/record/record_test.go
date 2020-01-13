package record

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRecord(t *testing.T) {
	record := NewRecord(1, "0.0.0.1")
	assert.Equal(t, UserID(1), record.UserID)
	assert.Equal(t, net.ParseIP("0.0.0.1").To4(), record.IP)
}
