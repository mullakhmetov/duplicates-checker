package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/internal/healthcheck"
)

type server struct {
	*Command
	srv *http.Server

	terminated chan struct{}
}

// Command starts http server
type Command struct {
	Port int `long:"port" env:"CHECKER_PORT" default:"8080" description:"port"`
	cmd.CommonOpts
}

// Execute command starts Rest server
func (c *Command) Execute(args []string) error {
	log.Printf("[INFO] start server on port %d", c.Port)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	server := c.newServer()
	err := server.run(ctx)
	if err != nil {
		log.Printf("[ERROR] terminated with error %+v", err)
		return err
	}

	log.Printf("[INFO] terminated")
	return nil
}

func (c *Command) newServer() *server {
	router := gin.Default()

	healthcheck.RegisterHandlers(router, c.CommonOpts.Revision)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: router,
	}

	return &server{
		Command:    c,
		srv:        srv,
		terminated: make(chan struct{}),
	}
}

func (s *server) run(ctx context.Context) error {
	go func() {
		// Graceful shutdown
		<-ctx.Done()
		s.srv.Shutdown(ctx)
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
