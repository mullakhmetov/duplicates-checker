package record

import "context"

const doubleLimit = 2

// Service encapsulates usecase logic
type Service interface {
	Create(ctx context.Context, record *Record) error
	IsDouble(ctx context.Context, u1, u2 UserID) (bool, error)
	Clear(ctx context.Context) error
}

type service struct {
	repo Repository
}

func (s *service) Create(ctx context.Context, record *Record) error {
	err := s.repo.Create(ctx, record)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) IsDouble(ctx context.Context, u1, u2 UserID) (bool, error) {
	u1IPs, err := s.repo.GetUserIPs(ctx, u1)
	if err != nil {
		return false, nil
	}
	u2IPs, err := s.repo.GetUserIPs(ctx, u2)
	if err != nil {
		return false, nil
	}

	return s.hasNCommons(u1IPs, u2IPs, doubleLimit), nil
}

func (s *service) Clear(ctx context.Context) error {
	return s.repo.Clean(ctx)
}

// Return the value true if `a` and `b` slices has n common values
func (s *service) hasNCommons(a, b []IP, n int) bool {
	temp := make(map[IP]int)

	for _, i := range a {
		// handle non-unique values
		temp[i] = 1
	}
	for _, i := range b {
		v, ok := temp[i]
		if ok && v < 2 {
			temp[i]++
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
