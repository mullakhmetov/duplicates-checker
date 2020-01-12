package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainServer(t *testing.T) {
	// schedule kill
	go func() {
		time.Sleep(2 * time.Second)
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.Nil(t, e)
	}()

	done := make(chan bool)
	port := chooseRandomUnusedPort()
	os.Args = []string{"test", "server", "--port=" + strconv.Itoa(port)}
	go func() {
		main()
		close(done)
	}()

	time.Sleep(time.Second)
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	<-done
}

func chooseRandomUnusedPort() (port int) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}
