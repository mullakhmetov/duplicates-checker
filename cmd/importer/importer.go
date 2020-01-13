package importer

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/jmoiron/sqlx"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/internal/record"
)

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
	pgDB   *sqlx.DB
}

func (s *sharedResources) Close() {
	log.Print("[INFO] closing shared resources")
	s.boltDB.Close()
	// s.pgDB.Close()
}

type importer struct {
	*Command

	*services
	*sharedResources
	random *rand.Rand

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

	// pgDB, err := record.NewPG()
	// if err != nil {
	// 	return nil, err
	// }
	// recordRepo, err := record.NewPGRepository(pgDB)
	// if err != nil {
	// 	return nil, err
	// }

	recordService := record.NewService(recordRepo)

	s := &importer{
		Command: c,
		services: &services{
			recordService: recordService,
		},
		sharedResources: &sharedResources{
			boltDB: boltDB,
			// pgDB:   pgDB,
		},
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
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

func (i *importer) addRecord(ctx context.Context, record *record.Record) error {
	return i.recordService.AddRecord(ctx, record)
}

func (imp *importer) createRecords(ctx context.Context, records []*record.Record) error {
	// for i := 0; i < len(records); i += 1e3 {
	// 	j := i + 1e3
	// 	if j > len(records) {
	// 		j = len(records)
	// 	}
	// 	if err := imp.recordService.CreateBatch(ctx, records[i:j]); err != nil {
	// 		return err
	// 	}
	// 	fmt.Println("commit: %d - %d", i, j)
	// }
	// return nil
	// file, err := os.Create("db.csv")
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// w := csv.NewWriter(file)
	// for _, record := range records {
	// 	if err := w.Write([]string{strconv.Itoa(int(record.UserID)), strconv.Itoa(int(net.IP))}); err != nil {
	// 		log.Fatalln("error writing record to csv:", err)
	// 	}
	// }
	// return nil
	batchSize := 1000
	for i := 0; i < len(records); i += batchSize {
		j := i + batchSize
		if j > len(records) {
			j = len(records)
		}
		// fmt.Println(records[i:j]) // Process the batch.
		if err := imp.recordService.BulkAddRecords(ctx, records[i:j]); err != nil {
			return err
		}
		fmt.Printf("process batch %d -- %d\n", i, j)
	}

	// for _, record := range records {
	// 	if err := imp.createRecord(ctx, record); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

type testLog struct {
	uID uint32
	IP  string
}

func (i *importer) generateDbg(ctx context.Context) (records []*record.Record, err error) {
	fmt.Println("GG DBG")
	// keep sync with cases in importer_test.go
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
		rec := record.NewRecord(record.UserID(log.uID), log.IP)
		if err != nil {
			return records, err
		}
		records = append(records, rec)
	}
	fmt.Printf("len: %d\n", len(records))

	return records, nil
}

func (imp *importer) generate(ctx context.Context) (err error) {
	if imp.dbg {
		records, _ := imp.generateDbg(ctx)
		return imp.createRecords(ctx, records)
	}
	getIP := ipsGetter()

	var i, ipsCount, reqCount uint
	var ips []net.IP
	var ip net.IP
	var rec *record.Record

	var JJ int

	superBatch := int(1e5)
	records := make([]*record.Record, 0, superBatch)

	for uID := record.UserID(1); uID < 1*1e5; uID++ {
		ipsCount = imp.getUserIPSCount(10)
		ips = make([]net.IP, 0, ipsCount)
		for i = uint(0); i <= ipsCount; i++ {
			ips = append(ips, getIP())
		}
		reqCount = imp.getUserRequestsCount(1 * 1e3)

		for i = uint(1); i <= reqCount; i++ {
			JJ++
			// fmt.Printf("%d\n", JJ)

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				ip = ips[i%ipsCount]
				rec = &record.Record{UserID: uID, IP: ip}
				if err != nil {
					return err
				}
				records = append(records, rec)
			}
		}
		if len(records) >= superBatch {
			if err = imp.createRecords(ctx, records); err != nil {
				return err
			}
			records = make([]*record.Record, 0, superBatch)
			fmt.Printf("superbatch loaded: %d\n", JJ)
		}
	}

	return imp.createRecords(ctx, records)
}

func (i *importer) Wait() {
	<-i.terminated
}

// Returns an exponentially distributed value from 1 to int(MaxFloat64) with `max` limit
// Represents how many different IPs used by user
func (i *importer) getUserIPSCount(max uint) uint {

	count := uint(i.random.ExpFloat64() + 1)
	if count > max {
		count = max
	}
	return count
}

// Returns normally distributed value from 1 to int(MaxFloat64) with `max` limit and 500 mean
// Represents how many requests user did
func (i *importer) getUserRequestsCount(max uint) (res uint) {

	desiredStdDev := 1.0
	desiredMean := 5 * 1e2
	res = uint(i.random.NormFloat64()*desiredStdDev + desiredMean)
	if res > max {
		res = max
	}
	return res
}

// ring over all possible IPs
func ipsGetter() func() net.IP {
	curr := uint32(0)
	return func() net.IP {
		if curr == 500 {
			curr = uint32(0)
		} else {
			curr++
		}
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, curr)
		return ip
	}
}
