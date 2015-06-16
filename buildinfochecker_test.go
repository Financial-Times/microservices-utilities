package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"strconv"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func newLocalListener(port int) net.Listener {
	customPort := strconv.Itoa(port)
	l, err := net.Listen("tcp", "127.0.0.1:"+customPort)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:"+customPort); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func newHttpTestServer(handler http.Handler, port int) *httptest.Server {
	return &httptest.Server{
		Listener: newLocalListener(port),
		Config:   &http.Server{Handler: handler},
	}
}

const (
	randomPort = 12221
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = newHttpTestServer(mux, randomPort)

	server.Start()
	fmt.Print("Running on ")
}

func teardown() {
	server.Close()
}

func TestParsingOfConfigFile(t *testing.T) {
	assert := assert.New(t)

	hostsAndPorts := ParseConfig("./test/services.yml")
	for _, hostPort := range hostsAndPorts {
		assert.Equal("localhost", hostPort.Hostname)
		assert.True(12223-hostPort.Port > 0 && 12223-hostPort.Port < 3)
	}
}

func TestAssertionOfVersion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/build-info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"groupId":"testing","artifactId":"testing-service","version":"1.2.3"}`)
	},
	)

	assert := assert.New(t)

	hostAndPort := HostAndPort{
		Hostname: "localhost",
		Port:     randomPort,
	}

	correctVersion, message := AssertVersion(hostAndPort, "1.2.3")
	assert.True(correctVersion)
	assert.Contains("Success", message)
}
