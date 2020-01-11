package record

import "context"

// Service encapsulates usecase logic
type Service interface {
	Create(ctx context.Context) error
	IsDouble(ctx context.Context, u1, u2 UserID) (bool, error)
	Clear(ctx context.Context) error
}

type service struct {
	repo Repository
}

func (s service) Create(ctx context.Context) error {
	return nil
}

func (s service) IsDouble(ctx context.Context, u1, u2 UserID) (bool, error) {
	return true, nil
}

func (s service) Clear(ctx context.Context) error {
	return nil
}

// NewService returns Service implementation
func NewService(repo Repository) Service {
	return service{repo}
}
