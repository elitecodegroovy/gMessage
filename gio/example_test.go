package gio_test

import (
	"fmt"
	"time"

	"github.com/elitecodegroovy/gmessage/gio"
)

// Shows different ways to create a Conn
func ExampleConnect() {

	nc, _ := gio.Connect(gio.DefaultURL)
	nc.Close()

	nc, _ = gio.Connect("nats://derek:secretpassword@demo.gio.io:6222")
	nc.Close()

	nc, _ = gio.Connect("tls://derek:secretpassword@demo.gio.io:4443")
	nc.Close()

	opts := gio.Options{
		AllowReconnect: true,
		MaxReconnect:   10,
		ReconnectWait:  5 * time.Second,
		Timeout:        1 * time.Second,
	}

	nc, _ = opts.Connect()
	nc.Close()
}

// This Example shows an asynchronous subscriber.
func ExampleConn_Subscribe() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	nc.Subscribe("foo", func(m *gio.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})
}

// This Example shows a synchronous subscriber.
func ExampleConn_SubscribeSync() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	m, err := sub.NextMsg(1 * time.Second)
	if err == nil {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	} else {
		fmt.Println("NextMsg timed out.")
	}
}

func ExampleSubscription_NextMsg() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	m, err := sub.NextMsg(1 * time.Second)
	if err == nil {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	} else {
		fmt.Println("NextMsg timed out.")
	}
}

func ExampleSubscription_Unsubscribe() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	// ...
	sub.Unsubscribe()
}

func ExampleConn_Publish() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	nc.Publish("foo", []byte("Hello World!"))
}

func ExampleConn_PublishMsg() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	msg := &gio.Msg{Subject: "foo", Reply: "bar", Data: []byte("Hello World!")}
	nc.PublishMsg(msg)
}

func ExampleConn_Flush() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	msg := &gio.Msg{Subject: "foo", Reply: "bar", Data: []byte("Hello World!")}
	for i := 0; i < 1000; i++ {
		nc.PublishMsg(msg)
	}
	err := nc.Flush()
	if err == nil {
		// Everything has been processed by the server for nc *Conn.
	}
}

func ExampleConn_FlushTimeout() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	msg := &gio.Msg{Subject: "foo", Reply: "bar", Data: []byte("Hello World!")}
	for i := 0; i < 1000; i++ {
		nc.PublishMsg(msg)
	}
	// Only wait for up to 1 second for Flush
	err := nc.FlushTimeout(1 * time.Second)
	if err == nil {
		// Everything has been processed by the server for nc *Conn.
	}
}

func ExampleConn_Request() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	nc.Subscribe("foo", func(m *gio.Msg) {
		nc.Publish(m.Reply, []byte("I will help you"))
	})
	nc.Request("foo", []byte("help"), 50*time.Millisecond)
}

func ExampleConn_QueueSubscribe() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	received := 0

	nc.QueueSubscribe("foo", "worker_group", func(_ *gio.Msg) {
		received++
	})
}

func ExampleSubscription_AutoUnsubscribe() {
	nc, _ := gio.Connect(gio.DefaultURL)
	defer nc.Close()

	received, wanted, total := 0, 10, 100

	sub, _ := nc.Subscribe("foo", func(_ *gio.Msg) {
		received++
	})
	sub.AutoUnsubscribe(wanted)

	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()

	fmt.Printf("Received = %d", received)
}

func ExampleConn_Close() {
	nc, _ := gio.Connect(gio.DefaultURL)
	nc.Close()
}

// Shows how to wrap a Conn into an EncodedConn
func ExampleNewEncodedConn() {
	nc, _ := gio.Connect(gio.DefaultURL)
	c, _ := gio.NewEncodedConn(nc, "json")
	c.Close()
}

// EncodedConn can publish virtually anything just
// by passing it in. The encoder will be used to properly
// encode the raw Go type
func ExampleEncodedConn_Publish() {
	nc, _ := gio.Connect(gio.DefaultURL)
	c, _ := gio.NewEncodedConn(nc, "json")
	defer c.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	me := &person{Name: "derek", Age: 22, Address: "85 Second St"}
	c.Publish("hello", me)
}

// EncodedConn's subscribers will automatically decode the
// wire data into the requested Go type using the Decode()
// method of the registered Encoder. The callback signature
// can also vary to include additional data, such as subject
// and reply subjects.
func ExampleEncodedConn_Subscribe() {
	nc, _ := gio.Connect(gio.DefaultURL)
	c, _ := gio.NewEncodedConn(nc, "json")
	defer c.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	c.Subscribe("hello", func(p *person) {
		fmt.Printf("Received a person! %+v\n", p)
	})

	c.Subscribe("hello", func(subj, reply string, p *person) {
		fmt.Printf("Received a person on subject %s! %+v\n", subj, p)
	})

	me := &person{Name: "derek", Age: 22, Address: "85 Second St"}
	c.Publish("hello", me)
}

// BindSendChan() allows binding of a Go channel to a gio
// subject for publish operations. The Encoder attached to the
// EncodedConn will be used for marshaling.
func ExampleEncodedConn_BindSendChan() {
	nc, _ := gio.Connect(gio.DefaultURL)
	c, _ := gio.NewEncodedConn(nc, "json")
	defer c.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	ch := make(chan *person)
	c.BindSendChan("hello", ch)

	me := &person{Name: "derek", Age: 22, Address: "85 Second St"}
	ch <- me
}

// BindRecvChan() allows binding of a Go channel to a gio
// subject for subscribe operations. The Encoder attached to the
// EncodedConn will be used for un-marshaling.
func ExampleEncodedConn_BindRecvChan() {
	nc, _ := gio.Connect(gio.DefaultURL)
	c, _ := gio.NewEncodedConn(nc, "json")
	defer c.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	ch := make(chan *person)
	c.BindRecvChan("hello", ch)

	me := &person{Name: "derek", Age: 22, Address: "85 Second St"}
	c.Publish("hello", me)

	// Receive the publish directly on a channel
	who := <-ch

	fmt.Printf("%v says hello!\n", who)
}
