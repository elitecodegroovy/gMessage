package main

import (
	"flag"
	"log"
	"runtime"

	"nats-io/go-nats"
	"time"
)

// NOTE: Use tls scheme for TLS, e.g. nats-rply -s tls://demo.nats.io:4443 foo hello
func usage() {
	log.Fatalf("Usage: nats-rply [-s server][-t] <subject> <response>\n")
}

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var showTime = flag.Bool("t", false, "Display timestamps")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()


	nc, err := nats.Connect(*urls)
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	subj, reply, i :=  "", "", 0
	if len(args) == 0 {
		subj = "NATS"
		reply = " NATS response "
	}else if len(args) < 2 {
		usage()
	}else {
		subj, reply, i = args[0], args[1], 0
	}

	//

	nc.Subscribe(subj, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
		currentTime := time.Now().Local()
		newFormatTime := currentTime.Format("2006-01-02 15:04:05.000")
		nc.Publish(msg.Reply, []byte(reply+ newFormatTime))
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]\n", subj)
	if *showTime {
		log.SetFlags(log.LstdFlags)
	}

	runtime.Goexit()
}
