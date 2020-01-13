package record

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const bucketName = "USER_INFO"

// Repository encapsulates the logic to access domain models
type Repository interface {
	GetUserInfo(ctx context.Context, userID UserID) (*UserInfo, error)
	AddRecord(ctx context.Context, record *Record) error
	BulkAddRecords(ctx context.Context, records []*Record) error
	Clean(ctx context.Context) error
}

// UserInfo contains user's info. UserInfo accumulates all user logs
type UserInfo struct {
	UserID UserID
	IPs    []net.IP
}

type boltIP uint32

type ipDecoder struct{}

func (d *ipDecoder) Encode(ip net.IP) boltIP {
	ip4 := ip.To4()
	return boltIP(binary.BigEndian.Uint32(ip4))
}

func (d *ipDecoder) Decode(ip boltIP) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(ip))
	return net.IP(b)
}

// bolt specific user's info structure. Stored in boltdb
type boltUserInfo struct {
	UserID UserID
	IPset  map[boltIP]bool
	*ipDecoder
}

func (bu *boltUserInfo) toUserInfo() *UserInfo {
	ips := make([]net.IP, 0, len(bu.IPset))
	for ip := range bu.IPset {
		ips = append(ips, bu.Decode(ip))
	}
	return &UserInfo{UserID: bu.UserID, IPs: ips}
}

type boltRepository struct {
	DB  *bolt.DB
	BKT string
}

type key []byte

func getKey(userID UserID) key {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(userID))

	return key
}

// Get returns UserInfo by UserID or nil if it doesn't exist
func (b *boltRepository) GetUserInfo(ctx context.Context, userID UserID) (*UserInfo, error) {
	userInfo := &UserInfo{}
	err := b.DB.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.BKT))

		v := bkt.Get(getKey(userID))
		if v == nil {
			return nil
		}

		boltUserInfo := boltUserInfo{}
		if err := json.Unmarshal(v, &boltUserInfo); err != nil {
			return err
		}
		userInfo = boltUserInfo.toUserInfo()

		return nil
	})

	return userInfo, err
}

// AddRecord does not add the record to storage literally. It gets user's UserInfo from storage
// and updates it with new info. UserInfo will be created if it doesn't exist yet.
func (b *boltRepository) AddRecord(ctx context.Context, record *Record) error {
	tx, err := b.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()

	return b.createOrUpdateBoltUserInfo(ctx, tx, record)
}

func (b *boltRepository) createBoltUserInfo(ctx context.Context, tx *bolt.Tx, record *Record) error {
	boltUserInfo := boltUserInfo{UserID: record.UserID}
	boltUserInfo.IPset = map[boltIP]bool{boltUserInfo.Encode(record.IP): true}

	buf, err := json.Marshal(&boltUserInfo)
	if err != nil {
		return err
	}
	bkt := tx.Bucket([]byte(b.BKT))

	bkt.Put(getKey(record.UserID), []byte(buf))
	return nil
}

// BulkAddRecords processes []*Record and updates UserInfo for each user's
func (b *boltRepository) BulkAddRecords(ctx context.Context, records []*Record) error {
	tx, err := b.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		b.createOrUpdateBoltUserInfo(ctx, tx, record)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Clean deletes bucket
func (b *boltRepository) Clean(ctx context.Context) error {
	err := b.DB.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(b.BKT))
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

// NewBoltDB returns boltb connection
func NewBoltDB(name string, options *bolt.Options) (*bolt.DB, error) {
	db, err := bolt.Open(name, 0600, options)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (b *boltRepository) createOrUpdateBoltUserInfo(ctx context.Context, tx *bolt.Tx, record *Record) error {
	bkt := tx.Bucket([]byte(b.BKT))
	v := bkt.Get([]byte(getKey(record.UserID)))

	if v == nil {
		return b.createBoltUserInfo(ctx, tx, record)
	}

	boltUserInfo := boltUserInfo{}
	if err := json.Unmarshal(v, &boltUserInfo); err != nil {
		return err
	}

	return b.updateBoltUserInfo(ctx, tx, &boltUserInfo, record)
}

func (b *boltRepository) updateBoltUserInfo(ctx context.Context, tx *bolt.Tx, boltUserInfo *boltUserInfo, record *Record) error {
	boltUserInfo.IPset[boltUserInfo.Encode(record.IP)] = true

	buf, err := json.Marshal(&boltUserInfo)
	if err != nil {
		return err
	}

	bkt := tx.Bucket([]byte(b.BKT))
	bkt.Put(getKey(record.UserID), []byte(buf))
	return nil
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
