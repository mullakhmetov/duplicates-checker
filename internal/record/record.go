package record

import (
	"encoding/binary"
	"errors"
	"net"
)

type IP uint32

type IpSerializer struct{}

type Record struct {
	ID     uint64
	UserID uint64
	IP     IP
}

func (_ *IpSerializer) Encode(s string) (IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return IP(binary.BigEndian.Uint32(ip)), nil
}

func (_ *IpSerializer) Decode(i IP) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	ip := net.IP(b)
	return ip.String()
}
