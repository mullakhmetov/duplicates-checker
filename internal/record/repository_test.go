package record

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

var testDb = "/tmp/test.db"

func TestBoltRepo_BucketCreated(t *testing.T) {
	_, b, teardown := prepBoltRepo(t)
	defer teardown()

	b.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		assert.NotNil(t, bkt)
		return nil
	})
}

func TestBoltRepo_AddRecord(t *testing.T) {
	r, b, teardown := prepBoltRepo(t)
	defer teardown()

	uID := UserID(1)
	record := NewRecord(uID, ("0.0.0.1"))
	err := r.AddRecord(context.Background(), record)
	assert.NoError(t, err)

	boltUserInfo := boltUserInfo{}
	err = b.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		v := bkt.Get(getKey(uID))
		assert.NotNil(t, v)
		err := json.Unmarshal(v, &boltUserInfo)
		if err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, uID, boltUserInfo.UserID)
	assert.True(t, boltUserInfo.IPset[1])
}

func TestBoltRepo_BulkAddRecords(t *testing.T) {
	r, _, teardown := prepBoltRepo(t)
	defer teardown()

	record1 := NewRecord(1, ("1.1.1.1"))
	record2 := NewRecord(1, ("1.1.1.1"))
	record3 := NewRecord(1, ("2.2.2.2"))
	record4 := NewRecord(2, ("2.2.2.2"))

	err := r.BulkAddRecords(context.Background(), []*Record{record1, record2, record3, record4})

	assert.NoError(t, err)
}

func TestBoltRepo_createOrUpdateBehavior(t *testing.T) {
	r, b, teardown := prepBoltRepo(t)
	defer teardown()

	// create duplicates by IP & UserID keys
	record := NewRecord(1, ("1.1.1.1"))
	err := r.AddRecord(context.Background(), record)
	assert.NoError(t, err)

	record = NewRecord(1, ("1.1.1.1"))
	err = r.AddRecord(context.Background(), record)
	assert.NoError(t, err)

	err = b.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		stats := bkt.Stats()
		assert.Equal(t, 1, stats.KeyN)
		return nil
	})
	assert.NoError(t, err)

	record = NewRecord(1, ("2.2.2.2"))
	err = r.AddRecord(context.Background(), record)
	assert.NoError(t, err)

	record = NewRecord(2, ("1.1.1.1"))
	err = r.AddRecord(context.Background(), record)
	assert.NoError(t, err)

	err = b.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		stats := bkt.Stats()
		assert.Equal(t, 2, stats.KeyN)
		return nil
	})
	assert.NoError(t, err)

}

func TestBoltRepo_GetUserInfo(t *testing.T) {
	r, _, teardown := prepBoltRepo(t)
	defer teardown()

	uID1 := UserID(1)
	uID2 := UserID(2)
	ip1 := ("1.1.1.1")
	ip2 := ("2.2.2.2")
	ctx := context.Background()

	rec := NewRecord(uID1, ip1)
	r.AddRecord(ctx, rec)

	rec = NewRecord(uID1, ip2)
	r.AddRecord(ctx, rec)

	rec = NewRecord(uID2, ip1)
	r.AddRecord(ctx, rec)

	info, err := r.GetUserInfo(ctx, uID1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(info.IPs), info.IPs)
	assert.ElementsMatch(t, []net.IP{net.ParseIP(ip1).To4(), net.ParseIP(ip2).To4()}, info.IPs)

	info, err = r.GetUserInfo(ctx, uID2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(info.IPs))
	assert.Equal(t, []net.IP{net.ParseIP(ip1).To4()}, info.IPs)

	info, err = r.GetUserInfo(ctx, UserID(10))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(info.IPs))
}

func TestBoltRepo_Clean(t *testing.T) {
	r, b, teardown := prepBoltRepo(t)
	defer teardown()

	uID1 := UserID(1)

	ctx := context.Background()
	rec := NewRecord(uID1, ("1.1.1.1"))
	r.AddRecord(ctx, rec)
	err := r.Clean(ctx)
	assert.NoError(t, err)

	b.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucketName))
		assert.Nil(t, bkt)
		return nil
	})
}

type recordTestCase struct {
	i boltIP
	s string
}

func TestRecordSuccEncode(t *testing.T) {
	d := ipDecoder{}

	succCases := getSuccCases()

	for _, c := range succCases {
		got := d.Encode(net.ParseIP(c.s).To4())
		assert.Equal(t, c.i, got)
	}
}

func TestRecordSuccDecode(t *testing.T) {
	d := ipDecoder{}

	succCases := getSuccCases()

	for _, c := range succCases {
		got := d.Decode(c.i)
		assert.Equal(t, net.ParseIP(c.s).To4(), got)
	}
}

func getSuccCases() []recordTestCase {
	return []recordTestCase{
		{0, ("0.0.0.0")},
		{16843009, ("1.1.1.1")},
		{4294967295, ("255.255.255.255")},
	}
}

func prepBoltRepo(t *testing.T) (repo Repository, bolt *bolt.DB, teardown func()) {
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
