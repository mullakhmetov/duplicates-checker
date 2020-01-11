package record

import (
	"context"

	"github.com/boltdb/bolt"
)

// Repository encapsulates the logic to access albums from the data source
type Repository interface {
	Get(ctx context.Context, id string) (Record, error)
	Create(ctx context.Context, album Record) error
	Clear(ctx context.Context) error
}

type boltRepository struct {
	db *bolt.DB
}

func (b *boltRepository) Get(ctx context.Context, id string) (Record, error) {
	return Record{}, nil
}

func (b *boltRepository) Create(ctx context.Context, album Record) error {
	return nil
}

func (b *boltRepository) Clear(ctx context.Context) error {
	return nil
}

// NewBoltRepository makes boltb Repository implementation
func NewBoltRepository(db *bolt.DB) Repository {
	return &boltRepository{db}
}
