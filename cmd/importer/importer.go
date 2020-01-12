package importer

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/internal/record"
)

// Command ...
type Command struct {
	cmd.CommonOpts
}

type services struct {
	recordService record.Service
}

type sharedResources struct {
	boltDB *bolt.DB
}

func (s *sharedResources) Close() {
	log.Print("[INFO] closing BoltDB")
	s.boltDB.Close()
}

type importer struct {
	*Command

	*services
	*sharedResources

	dbg        bool
	terminated chan struct{}
}

// Execute command starts importing data to store
func (c *Command) Execute(args []string) error {
	log.Printf("[INFO] start import records process. Debug mode: %t", c.Dbg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	importer, err := c.newImporter(c.Dbg)
	if err != nil {
		return err
	}
	err = importer.run(ctx)
	if err != nil {
		log.Printf("[ERROR] terminated with error %+v", err)
		return err
	}

	log.Printf("[INFO] terminated")
	return nil
}

func (c *Command) newImporter(dbg bool) (*importer, error) {
	boltDB, err := record.NewBoltDB(c.BoltDBName, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	recordRepo, err := record.NewBoltRepository(boltDB)
	if err != nil {
		return nil, err
	}
	recordService := record.NewService(recordRepo)

	s := &importer{
		Command: c,
		services: &services{
			recordService: recordService,
		},
		sharedResources: &sharedResources{
			boltDB: boltDB,
		},
		dbg:        dbg,
		terminated: make(chan struct{}),
	}
	return s, nil
}

func (i *importer) run(ctx context.Context) error {
	go func() {
		// Graceful shutdown
		<-ctx.Done()
		i.sharedResources.Close()
		log.Print("[INFO] importer was shut down")
	}()

	if err := i.generate(ctx); err != nil {
		return err
	}

	i.sharedResources.Close()
	close(i.terminated)
	return nil
}

func (i *importer) createRecord(ctx context.Context, record *record.Record) error {
	return i.recordService.Create(ctx, record)
}

type testLog struct {
	uID uint32
	IP  string
}

func (i *importer) generateDbg(ctx context.Context) error {

	logs := []testLog{
		testLog{1, "127.0.0.1"},
		testLog{1, "127.0.0.2"},
		testLog{2, "127.0.0.1"},
		testLog{2, "127.0.0.2"},
		testLog{2, "127.0.0.3"},
		testLog{3, "127.0.0.3"},
		testLog{3, "127.0.0.1"},
		testLog{4, "127.0.0.1"},
	}

	for _, log := range logs {
		rec, err := record.NewRecord(log.IP, record.UserID(log.uID))
		if err != nil {
			return err
		}
		if err := i.createRecord(ctx, rec); err != nil {
			return err
		}
	}

	return nil
}

func (i *importer) generate(ctx context.Context) error {
	if i.dbg {
		return i.generateDbg(ctx)
	}
	return nil
}

func (i *importer) Wait() {
	<-i.terminated
}

// Returns an exponentially distributed value from 1 to int(MaxFloat64) with `max` limit
// Represents how many different IPs used by user
func (c *Command) getUserIPSCount(max uint) uint {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := uint(r.ExpFloat64() + 1)
	if count > max {
		count = max
	}
	return count
}

// Returns normally distributed value from 1 to int(MaxFloat64) with `max` limit and 500k mean
// Represents how many requests user did
func (c *Command) getUserRequestsCount(max uint) uint {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	desiredStdDev := 1.0
	desiredMean := 5 * 1e5
	return uint(r.NormFloat64()*desiredStdDev + desiredMean)
}

// ring over all possible IPs
func ipsGetter() func() record.IP {
	curr := record.IP(0)
	return func() record.IP {
		if curr == record.MaxIP {
			curr = record.IP(0)
		} else {
			curr++
		}
		return curr
	}
}

// func (c *Command) generateRecords(users int, ) error {

// }
