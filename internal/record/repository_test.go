package record

import (
	"context"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

var testDb = "/tmp/test.db"

func TestBoltRepo_BucketCreated(t *testing.T) {
	_, b, teardown := prepRepo(t)
	defer teardown()

	b.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		assert.NotNil(t, bkt)
		return nil
	})
}

func TestBoltRepo_Create(t *testing.T) {
	r, _, teardown := prepRepo(t)
	defer teardown()

	record, err := NewRecord("1.1.1.1", 1)
	assert.NoError(t, err)
	err = r.Create(context.Background(), record)

	assert.NoError(t, err)
}

func TestBoltRepo_createOrUpdate(t *testing.T) {
	r, b, teardown := prepRepo(t)
	defer teardown()

	// create duplicates by IP & UserID keys
	record, err := NewRecord("1.1.1.1", 1)
	assert.NoError(t, err)
	err = r.Create(context.Background(), record)
	assert.NoError(t, err)

	record, err = NewRecord("1.1.1.1", 1)
	assert.NoError(t, err)
	err = r.Create(context.Background(), record)
	assert.NoError(t, err)

	err = b.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		stats := bkt.Stats()
		assert.Equal(t, 1, stats.KeyN)
		return nil
	})
	assert.NoError(t, err)

	record, err = NewRecord("2.2.2.2", 1)
	assert.NoError(t, err)
	err = r.Create(context.Background(), record)
	assert.NoError(t, err)

	record, err = NewRecord("1.1.1.1", 2)
	assert.NoError(t, err)
	err = r.Create(context.Background(), record)
	assert.NoError(t, err)

	err = b.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		stats := bkt.Stats()
		assert.Equal(t, 3, stats.KeyN)
		return nil
	})
	assert.NoError(t, err)

}

func TestBoltRepo_GetUserIPs(t *testing.T) {
	r, _, teardown := prepRepo(t)
	defer teardown()

	uID1 := UserID(1)
	uID2 := UserID(2)
	ctx := context.Background()

	rec, err := NewRecord("1.1.1.1", uID1)
	assert.NoError(t, err)
	r.Create(ctx, rec)

	rec, err = NewRecord("2.2.2.2", uID1)
	assert.NoError(t, err)
	r.Create(ctx, rec)

	rec, err = NewRecord("1.1.1.1", uID2)
	assert.NoError(t, err)
	r.Create(ctx, rec)

	ips, err := r.GetUserIPs(ctx, uID1)
	assert.NoError(t, err)
	assert.Equal(t, len(ips), 2)

	ips, err = r.GetUserIPs(ctx, uID2)
	assert.NoError(t, err)
	assert.Equal(t, len(ips), 1)

	ips, err = r.GetUserIPs(ctx, UserID(10))
	assert.NoError(t, err)
	assert.Equal(t, len(ips), 0)
}

func TestBoltRepo_Clean(t *testing.T) {
	r, b, teardown := prepRepo(t)
	defer teardown()

	uID1 := UserID(1)

	ctx := context.Background()
	rec, err := NewRecord("1.1.1.1", uID1)
	assert.NoError(t, err)
	r.Create(ctx, rec)
	err = r.Clean(ctx)
	assert.NoError(t, err)

	b.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		assert.Nil(t, bkt)
		return nil
	})
}

func prepRepo(t *testing.T) (repo Repository, bolt *bolt.DB, teardown func()) {
	_ = os.Remove(testDb)

	bolt, err := NewBoltDB(testDb, nil)
	assert.NoError(t, err)

	repo, err = NewBoltRepository(bolt)

	teardown = func() {
		assert.NoError(t, bolt.Close())
		_ = os.Remove(testDb)
	}

	return repo, bolt, teardown
}
