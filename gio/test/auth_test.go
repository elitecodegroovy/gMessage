package test

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elitecodegroovy/gmessage/gio"
	"github.com/elitecodegroovy/gmessage/server"
	"github.com/elitecodegroovy/gmessage/test"
)

func TestAuth(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Username = "derek"
	opts.Password = "foo"
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	_, err := gio.Connect("nats://localhost:8232")
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}

	// This test may be a bit too strict for the future, but for now makes
	// sure that we correctly process the -ERR content on connect.
	if err.Error() != gio.ErrAuthorization.Error() {
		t.Fatalf("Expected error '%v', got '%v'", gio.ErrAuthorization, err)
	}

	nc, err := gio.Connect("nats://derek:foo@localhost:8232")
	if err != nil {
		t.Fatal("Should have connected successfully with a token")
	}
	nc.Close()

	// Use Options
	nc, err = gio.Connect("nats://localhost:8232", gio.UserInfo("derek", "foo"))
	if err != nil {
		t.Fatalf("Should have connected successfully with a token: %v", err)
	}
	nc.Close()
	// Verify that credentials in URL take precedence.
	nc, err = gio.Connect("nats://derek:foo@localhost:8232", gio.UserInfo("foo", "bar"))
	if err != nil {
		t.Fatalf("Should have connected successfully with a token: %v", err)
	}
	nc.Close()
}

func TestAuthFailNoDisconnectCB(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Username = "derek"
	opts.Password = "foo"
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	copts := gio.GetDefaultOptions()
	copts.Url = "nats://localhost:8232"
	receivedDisconnectCB := int32(0)
	copts.DisconnectedCB = func(nc *gio.Conn) {
		atomic.AddInt32(&receivedDisconnectCB, 1)
	}

	_, err := copts.Connect()
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}
	if atomic.LoadInt32(&receivedDisconnectCB) > 0 {
		t.Fatal("Should not have received a disconnect callback on auth failure")
	}
}

func TestAuthFailAllowReconnect(t *testing.T) {
	ts := RunServerOnPort(23232)
	defer ts.Shutdown()

	var servers = []string{
		"nats://localhost:23232",
		"nats://localhost:23233",
		"nats://localhost:23234",
	}

	ots2 := test.DefaultTestOptions
	ots2.Port = 23233
	ots2.Username = "ivan"
	ots2.Password = "foo"
	ts2 := RunServerWithOptions(ots2)
	defer ts2.Shutdown()

	ts3 := RunServerOnPort(23234)
	defer ts3.Shutdown()

	reconnectch := make(chan bool)

	opts := gio.GetDefaultOptions()
	opts.Servers = servers
	opts.AllowReconnect = true
	opts.NoRandomize = true
	opts.MaxReconnect = 10
	opts.ReconnectWait = 100 * time.Millisecond

	opts.ReconnectedCB = func(_ *gio.Conn) {
		reconnectch <- true
	}

	// Connect
	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Should have connected ok: %v", err)
	}
	defer nc.Close()

	// Stop the server
	ts.Shutdown()

	// The client will try to connect to the second server, and that
	// should fail. It should then try to connect to the third and succeed.

	// Wait for the reconnect CB.
	if e := Wait(reconnectch); e != nil {
		t.Fatal("Reconnect callback should have been triggered")
	}

	if nc.IsClosed() {
		t.Fatal("Should have reconnected")
	}

	if nc.ConnectedUrl() != servers[2] {
		t.Fatalf("Should have reconnected to %s, reconnected to %s instead", servers[2], nc.ConnectedUrl())
	}
}

func TestTokenAuth(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	secret := "S3Cr3T0k3n!"
	opts.Authorization = secret
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	_, err := gio.Connect("nats://localhost:8232")
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}

	tokenURL := fmt.Sprintf("nats://%s@localhost:8232", secret)
	nc, err := gio.Connect(tokenURL)
	if err != nil {
		t.Fatal("Should have connected successfully")
	}
	nc.Close()

	// Use Options
	nc, err = gio.Connect("nats://localhost:8232", gio.Token(secret))
	if err != nil {
		t.Fatalf("Should have connected successfully: %v", err)
	}
	nc.Close()
	// Verify that token in the URL takes precedence.
	nc, err = gio.Connect(tokenURL, gio.Token("badtoken"))
	if err != nil {
		t.Fatalf("Should have connected successfully: %v", err)
	}
	nc.Close()
}

func TestPermViolation(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Users = []*server.User{
		&server.User{
			Username: "ivan",
			Password: "pwd",
			Permissions: &server.Permissions{
				Publish:   []string{"foo"},
				Subscribe: []string{"bar"},
			},
		},
	}
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	errCh := make(chan error, 2)
	errCB := func(_ *gio.Conn, _ *gio.Subscription, err error) {
		errCh <- err
	}
	nc, err := gio.Connect(
		fmt.Sprintf("nats://ivan:pwd@localhost:%d", opts.Port),
		gio.ErrorHandler(errCB))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer nc.Close()

	// Cause a publish error
	nc.Publish("bar", []byte("fail"))
	// Cause a subscribe error
	nc.Subscribe("foo", func(_ *gio.Msg) {})

	expectedErrorTypes := []string{"publish", "subscription"}
	for _, expectedErr := range expectedErrorTypes {
		select {
		case e := <-errCh:
			if !strings.Contains(e.Error(), gio.PERMISSIONS_ERR) {
				t.Fatalf("Did not receive error about permissions")
			}
			if !strings.Contains(e.Error(), expectedErr) {
				t.Fatalf("Did not receive error about %q, got %v", expectedErr, e.Error())
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Did not get the permission error")
		}
	}
	// Make sure connection has not been closed
	if nc.IsClosed() {
		t.Fatal("Connection should be not be closed")
	}
}
