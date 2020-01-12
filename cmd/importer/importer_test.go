package importer

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/internal/record"
	"github.com/stretchr/testify/assert"
)

var testDb = "/tmp/test.db"

type importerTestCases struct {
	u1ID record.UserID
	u2ID record.UserID

	res bool
}

var cases = []importerTestCases{
	importerTestCases{record.UserID(1), record.UserID(2), true},
	importerTestCases{record.UserID(1), record.UserID(3), false},
	importerTestCases{record.UserID(2), record.UserID(1), true},
	importerTestCases{record.UserID(2), record.UserID(3), true},
	importerTestCases{record.UserID(3), record.UserID(2), true},
	importerTestCases{record.UserID(1), record.UserID(4), false},
	importerTestCases{record.UserID(3), record.UserID(1), false},
	importerTestCases{record.UserID(1), record.UserID(1), true},
}

func TestImorter(t *testing.T) {
	_ = os.Remove(testDb)

	ctx := context.Background()
	c := Command{CommonOpts: cmd.CommonOpts{BoltDBName: testDb}}
	i, err := c.newImporter(true)
	assert.NoError(t, err)

	err = i.generateDbg(ctx)
	assert.NoError(t, err)

	for _, c := range cases {
		res, err := i.recordService.IsDouble(ctx, c.u1ID, c.u2ID)
		assert.NoError(t, err)
		assert.Equal(t, c.res, res, fmt.Sprintf("%v", c))
	}
}
