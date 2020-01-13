package importer

import (
	"context"
	"fmt"
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

var batchSize = int(1e5)

// Command ...
type Command struct {
	usersCountLimit     int
	requestPerUserLimit int `long:"port" env:"CHECKER_PORT" default:"8080" description:"port"`
	ipsPerUserLimit     int

	cmd.CommonOpts
}

type services struct {
	recordService record.Service
}

type sharedResources struct {
	boltDB *bolt.DB
}

func (s *sharedResources) Close() {
	log.Print("[INFO] closing shared resources")
	s.boltDB.Close()
}

type importer struct {
	*Command

	*services
	*sharedResources

	gen        *generator
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
		gen:        &generator{rand.New(rand.NewSource(time.Now().UnixNano()))},
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

	var ch chan *record.Record
	if i.dbg {
		ch = i.gen.generateDbg(ctx)
	} else {
		ch = i.gen.generate(ctx)
	}

	records := make([]*record.Record, 0, batchSize)
	C := 1
	for rec := range ch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			records = append(records, rec)
			if len(records) >= batchSize {
				if err := i.recordService.BulkAddRecords(ctx, records); err != nil {
					return err
				}
				fmt.Printf("%d records loaded\n", C)
				records = make([]*record.Record, 0, batchSize)
			}
			C++
		}
	}

	if len(records) > 0 {
		if err := i.recordService.BulkAddRecords(ctx, records); err != nil {
			return err
		}
	}

	i.sharedResources.Close()
	close(i.terminated)
	return nil
}

func (i *importer) Wait() {
	<-i.terminated
}
