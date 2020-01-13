package record

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

const doubleLimit = 2

// Service encapsulates usecase logic
type Service interface {
	AddRecord(ctx context.Context, record *Record) error
	BulkAddRecords(ctx context.Context, records []*Record) error
	IsDuple(ctx context.Context, u1, u2 UserID) (bool, error)
	Clear(ctx context.Context) error
}

type service struct {
	repo Repository
}

// AddRecord processes new record
func (s *service) AddRecord(ctx context.Context, record *Record) error {
	err := s.repo.AddRecord(ctx, record)
	if err != nil {
		return err
	}
	return nil
}

// BulkAddRecords processes new records
func (s *service) BulkAddRecords(ctx context.Context, records []*Record) error {
	err := s.repo.BulkAddRecords(ctx, records)
	if err != nil {
		return err
	}
	return nil
}

// IsDuple returns true if users are duplicates
func (s *service) IsDuple(ctx context.Context, u1, u2 UserID) (bool, error) {
	u1Info, err := s.repo.GetUserInfo(ctx, u1)
	if err != nil {
		return false, err
	}
	u2Info, err := s.repo.GetUserInfo(ctx, u2)
	if err != nil {
		return false, err
	}
	fmt.Printf("u1: %v u2: %v\n", u1Info, u2Info)
	return s.hasNCommons(u1Info.IPs, u2Info.IPs, doubleLimit), nil
}

func (s *service) Clear(ctx context.Context) error {
	return s.repo.Clean(ctx)
}

// Return the value true if `a` and `b` slices has n common values
func (s *service) hasNCommons(a, b []net.IP, n int) bool {
	temp := make(map[uint32]int)

	for _, i := range a {
		// handle non-unique values
		temp[uint32(binary.BigEndian.Uint32(i.To4()))] = 1
	}
	for _, i := range b {
		k := uint32(binary.BigEndian.Uint32(i.To4()))
		v, ok := temp[k]
		if ok && v < 2 {
			temp[k]++
		}
	}

	var commons int
	for _, v := range temp {
		// v presents in both slices
		if v >= 2 {
			commons++
		}
		if commons >= n {
			return true
		}
	}
	return false
}

// NewService returns Service implementation
func NewService(repo Repository) Service {
	return &service{repo}
}
