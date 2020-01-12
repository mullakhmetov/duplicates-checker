package record

import (
	"context"
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const bucketName = "RECORD"

// Repository encapsulates the logic to access albums from the data source
type Repository interface {
	GetUserIPs(ctx context.Context, userID UserID) ([]IP, error)
	Create(ctx context.Context, record *Record) (ID, error)
	Clean(ctx context.Context) error
}

type boltRepository struct {
	DB  *bolt.DB
	BKT string
}

func (b *boltRepository) GetUserIPs(ctx context.Context, userID UserID) (ips []IP, err error) {

	ipSet := make(map[IP]bool)

	err = b.DB.View(func(tx *bolt.Tx) error {
		var e error
		bkt := tx.Bucket([]byte(b.BKT))

		bkt.ForEach(func(k, v []byte) error {
			r := Record{}
			if e = json.Unmarshal(v, &r); e != nil {
				return errors.Wrap(e, "failed to unmarshal")
			}
			if r.UserID == userID {
				ipSet[r.IP] = true
			}
			return nil
		})
		return nil
	})

	ips = make([]IP, 0, len(ipSet))
	for ip := range ipSet {
		ips = append(ips, ip)
	}
	return ips, err
}

func (b *boltRepository) Create(ctx context.Context, record *Record) (ID, error) {
	err := b.DB.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.BKT))
		id, _ := bkt.NextSequence()
		record.ID = ID(id)

		buf, err := json.Marshal(record)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(itob(uint64(record.ID))), []byte(buf))
	})
	return record.ID, err
}

func (b *boltRepository) Clean(ctx context.Context) error {
	err := b.DB.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(b.BKT))
	})
	return err
}

func (b *boltRepository) createBucketIfNotExists(bkt string) error {
	err := b.DB.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists([]byte(b.BKT)); e != nil {
			return errors.Wrapf(e, "failed to create bucket %s", b.BKT)
		}
		return nil
	})
	return err
}

// NewBoltRepository makes boltb Repository implementation, creates a bucket if it doesn't exist
func NewBoltRepository(db *bolt.DB) (Repository, error) {
	r := boltRepository{db, bucketName}
	err := r.createBucketIfNotExists(bucketName)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// NewBoltDB return boltb connection
func NewBoltDB(name string, options *bolt.Options) (*bolt.DB, error) {
	db, err := bolt.Open(name, 0600, options)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
