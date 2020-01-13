package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/internal/healthcheck"
	"github.com/mullakhmetov/duplicates-checker/internal/record"
)

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

type server struct {
	*Command
	srv *http.Server

	*services
	*sharedResources

	terminated chan struct{}
}

// Command starts http server
type Command struct {
	Port int `long:"port" env:"CHECKER_PORT" default:"8080" description:"port"`
	cmd.CommonOpts
}

// Execute command starts Rest server
func (c *Command) Execute(args []string) error {
	log.Printf("[INFO] start server on port %d. Debug mode: %t", c.Port, c.Dbg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	server, err := c.newServer()
	if err != nil {
		return err
	}
	err = server.run(ctx)
	if err != nil {
		log.Printf("[ERROR] terminated with error %+v", err)
		return err
	}

	log.Printf("[INFO] terminated")
	return nil
}

func (c *Command) newServer() (*server, error) {
	router := gin.Default()
	if !c.Dbg {
		gin.SetMode("release")
	}

	healthcheck.RegisterHandlers(router, c.Revision)

	boltDB, err := record.NewBoltDB(c.BoltDBName, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	recordRepo, err := record.NewBoltRepository(boltDB)
	if err != nil {
		return nil, err
	}

	recordService := record.NewService(recordRepo)
	record.RegisterHandlers(router, recordService)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: router,
	}

	s := &server{
		Command: c,
		srv:     srv,
		services: &services{
			recordService: recordService,
		},
		sharedResources: &sharedResources{
			boltDB: boltDB,
		},
		terminated: make(chan struct{}),
	}
	return s, nil
}

func (s *server) run(ctx context.Context) error {
	go func() {
		// Graceful shutdown
		<-ctx.Done()
		s.srv.Shutdown(ctx)
		s.sharedResources.Close()
		log.Print("[INFO] server was shut down")
	}()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	close(s.terminated)
	return nil
}

func (s *server) Wait() {
	<-s.terminated
}
