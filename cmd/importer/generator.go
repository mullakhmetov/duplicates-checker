package importer

import (
	"context"
	"encoding/binary"
	"math/rand"
	"net"

	"github.com/mullakhmetov/duplicates-checker/internal/record"
)

type dbgRecord struct {
	uID uint32
	IP  string
}

type generator struct {
	random *rand.Rand
}

func (g *generator) generateDbg(ctx context.Context) chan *record.Record {
	logs := []dbgRecord{
		dbgRecord{1, "127.0.0.1"},
		dbgRecord{1, "127.0.0.2"},
		dbgRecord{2, "127.0.0.1"},
		dbgRecord{2, "127.0.0.2"},
		dbgRecord{2, "127.0.0.3"},
		dbgRecord{3, "127.0.0.3"},
		dbgRecord{3, "127.0.0.1"},
		dbgRecord{4, "127.0.0.1"},
	}

	ch := make(chan *record.Record)
	go func() {
		defer close(ch)

		for _, log := range logs {
			select {
			case ch <- record.NewRecord(record.UserID(log.uID), log.IP):
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	return ch
}

func (g *generator) generate(ctx context.Context, usersCount, requestsLimit, requestsMean, ipsLimit uint) chan *record.Record {
	getIP := ipsGetter()

	var i, ipsCount, reqCount uint
	var ips []string

	ch := make(chan *record.Record)
	go func() {
		defer close(ch)

		for uID := record.UserID(1); uID < record.UserID(usersCount); uID++ {
			ipsCount = g.getUserIPSCount(ipsLimit)
			ips = make([]string, 0, ipsCount)
			for i = uint(0); i <= ipsCount; i++ {
				ips = append(ips, getIP())
			}
			reqCount = g.getUserRequestsCount(requestsMean, requestsLimit)

			for i = uint(1); i <= reqCount; i++ {
				select {
				case <-ctx.Done():
					close(ch)
					return
				case ch <- record.NewRecord(uID, ips[i%ipsCount]):
				}
			}
		}
	}()
	return ch
}

// Returns an exponentially distributed value from 1 to int(MaxFloat64) with `max` limit
// Represents how many different IPs used by user
func (g *generator) getUserIPSCount(max uint) uint {

	count := uint(g.random.ExpFloat64() + 1)
	if count > max {
		count = max
	}
	return count
}

// Returns normally distributed value from 1 to int(MaxFloat64) with `max` limit
// Represents how many requests user did
func (g *generator) getUserRequestsCount(mean, max uint) (res uint) {
	desiredStdDev := 1.0
	res = uint(g.random.NormFloat64()*desiredStdDev + float64(mean))
	if res > max {
		res = max
	}
	return res
}

// ring over all possible IPs
func ipsGetter() func() string {
	curr := uint32(1)
	return func() string {
		if curr == 500 {
			curr = uint32(0)
		} else {
			curr++
		}
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, curr)
		return ip.String()
	}
}
