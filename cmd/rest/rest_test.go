package rest

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRest(t *testing.T) {
	port := chooseRandomUnusedPort()
	c := newCommand(port)
	server, err := c.newServer()
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		server.run(ctx)
		waitForHTTPServerStart(port)
	}()

	// healthcheck
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "pong"))

	cancel()
	server.Wait()
}

func TestRest_Shutdown(t *testing.T) {
	port := chooseRandomUnusedPort()
	c := newCommand(port)
	server, err := c.newServer()
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, func() {
		cancel()
	})

	err = server.run(ctx)
	assert.NoError(t, err)
	server.Wait()
}

func TestRest_Signal(t *testing.T) {
	done := make(chan struct{})
	go func() {
		<-done
		time.Sleep(250 * time.Millisecond)
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, err)
	}()

	port := chooseRandomUnusedPort()
	c := newCommand(port)

	p := flags.NewParser(c, flags.Default)
	args := []string{"--port=" + strconv.Itoa(port)}
	_, err := p.ParseArgs(args)
	require.NoError(t, err)
	close(done)
	err = c.Execute(args)
	assert.NoError(t, err, "execute should be without errors")
}

func chooseRandomUnusedPort() (port int) {
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}

func waitForHTTPServerStart(port int) {
	// wait for up to 3 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 300; i++ {
		time.Sleep(time.Millisecond * 10)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}

func newCommand(port int) *Command {
	return &Command{Port: port, CommonOpts: cmd.CommonOpts{BoltDBName: "test.db"}}
}
