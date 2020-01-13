package record

import (
	"net"
)

// UserID ...
type UserID uint32

// Record is the main domain entity represents each log record
type Record struct {
	UserID UserID
	IP     net.IP
}

// NewRecord creates Record by UserID and string IP
func NewRecord(id UserID, ips string) *Record {
	ip := net.ParseIP(ips).To4()
	return &Record{id, ip}
}
