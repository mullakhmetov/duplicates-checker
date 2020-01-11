package store

import "github.com/boltdb/bolt"

// NewBoltDB return boltb connection
func NewBoltDB(name string, options *bolt.Options) (*bolt.DB, error) {
	db, err := bolt.Open(name, 0600, options)
	if err != nil {
		return nil, err
	}
	return db, nil
}
