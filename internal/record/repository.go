package record

import "context"

// Repository encapsulates the logic to access albums from the data source
type Repository interface {
	Get(ctx context.Context, id string) (Record, error)
	Create(ctx context.Context, album Record) error
	Clear(ctx context.Context) error
}
