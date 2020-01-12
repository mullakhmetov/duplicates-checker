package record

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

// MaxIP is a max possible IP value
const MaxIP = IP(math.MaxUint32)

// ID specific type
type ID uint64

// IP specific type
type IP uint32

// UserID ...
type UserID uint64

// Record is the main domain entity represents each log record
type Record struct {
	ID     ID
	UserID UserID
	IP     IP
}

// NewRecord creates Record by string IP and UserID
func NewRecord(s string, id UserID) (*Record, error) {
	ser := IPSerializer{}
	ip, err := ser.Encode(s)
	if err != nil {
		return nil, err
	}
	return &Record{UserID: id, IP: ip}, nil
}

// IPSerializer serves to IP encoding/decoding
type IPSerializer struct{}

// Encode transforms string IP representation to uint32
func (i *IPSerializer) Encode(sIP string) (IP, error) {
	ip := net.ParseIP(sIP)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return IP(binary.BigEndian.Uint32(ip)), nil
}

// Decode transforms uint32 IP representation to string
func (i *IPSerializer) Decode(iIP IP) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(iIP))
	ip := net.IP(b)
	return ip.String()
}
