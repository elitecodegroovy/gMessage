package test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/elitecodegroovy/gmessage/gio"
	"github.com/elitecodegroovy/gmessage/server"

	gnatsd "github.com/elitecodegroovy/gmessage/test"
)

// So that we can pass tests and benchmarks...
type tLogger interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// TestLogger
type TestLogger tLogger

// Dumb wait program to sync on callbacks, etc... Will timeout
func Wait(ch chan bool) error {
	return WaitTime(ch, 5*time.Second)
}

// Wait for a chan with a timeout.
func WaitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

func stackFatalf(t tLogger, f string, args ...interface{}) {
	lines := make([]string, 0, 32)
	msg := fmt.Sprintf(f, args...)
	lines = append(lines, msg)

	// Generate the Stack of callers: Skip us and verify* frames.
	for i := 1; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		msg := fmt.Sprintf("%d - %s:%d", i, file, line)
		lines = append(lines, msg)
	}
	t.Fatalf("%s", strings.Join(lines, "\n"))
}

////////////////////////////////////////////////////////////////////////////////
// Creating client connections
////////////////////////////////////////////////////////////////////////////////

// NewDefaultConnection
func NewDefaultConnection(t tLogger) *gio.Conn {
	return NewConnection(t, gio.DefaultPort)
}

// NewConnection forms connection on a given port.
func NewConnection(t tLogger, port int) *gio.Conn {
	url := fmt.Sprintf("gio://localhost:%d", port)
	nc, err := gio.Connect(url)
	if err != nil {
		t.Fatalf("Failed to create default connection: %v\n", err)
		return nil
	}
	return nc
}

// NewEConn
func NewEConn(t tLogger) *gio.EncodedConn {
	ec, err := gio.NewEncodedConn(NewDefaultConnection(t), gio.DEFAULT_ENCODER)
	if err != nil {
		t.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

////////////////////////////////////////////////////////////////////////////////
// Running ggiod server in separate Go routines
////////////////////////////////////////////////////////////////////////////////

// RunDefaultServer will run a server on the default port.
func RunDefaultServer() *server.Server {
	return RunServerOnPort(gio.DefaultPort)
}

// RunServerOnPort will run a server on the given port.
func RunServerOnPort(port int) *server.Server {
	opts := gnatsd.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(opts)
}

// RunServerWithOptions will run a server with the given options.
func RunServerWithOptions(opts server.Options) *server.Server {
	return gnatsd.RunServer(&opts)
}

// RunServerWithConfig will run a server with the given configuration file.
func RunServerWithConfig(configFile string) (*server.Server, *server.Options) {
	return gnatsd.RunServerWithConfig(configFile)
}
