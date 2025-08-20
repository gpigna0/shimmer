package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/gpigna0/shimmer/util"
)

var SOCK = path.Join(os.Getenv("XDG_RUNTIME_DIR"), "shimmer.sock")

type publisher struct {
	conns   map[net.Conn]bool
	pubChan chan string
	*sync.Mutex
}

func newPublisher() publisher {
	var s sync.Mutex
	pub := publisher{
		make(map[net.Conn]bool),
		make(chan string),
		&s,
	}

	go func() {
		for v := range pub.pubChan {
			log.Println(v)
			go pub.broadcast(v)
		}
	}()

	return pub
}

func (p publisher) add(c net.Conn) {
	p.Lock()
	defer p.Unlock()
	p.conns[c] = true
}

func (p publisher) close() {
	p.Lock()
	defer p.Unlock()
	for k := range p.conns {
		delete(p.conns, k)
		k.Close()
	}
	close(p.pubChan)
}

func (p publisher) broadcast(msg string) {
	p.Lock()
	defer p.Unlock()
	for k := range p.conns {
		if _, err := fmt.Fprintln(k, msg); err != nil {
			log.Println("err:", err)
			delete(p.conns, k) // assume the client has disconnected
		}
	}
}

func pubBrightness(devName string, pubChan chan<- string) {
	if d, ok := util.Conf.Devices[devName]; ok {
		v, err := util.ReadFloat64(d.Path)
		if err != nil {
			v = -1
		}

		msg := fmt.Sprintf(
			"BRIGHTNESS::%s::%s::%s",
			devName,
			strconv.Itoa(int(v)),
			strconv.FormatFloat(util.ToPercent(v, d.Max, 0), 'f', -1, 64),
		)

		pubChan <- msg
	}
}

func handleConn(conn net.Conn, pub publisher, auto chan string, ctx context.Context, cancel context.CancelFunc) bool {
	s := bufio.NewScanner(conn)

LOOP:
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			break LOOP
		default:
			if !s.Scan() {
				break LOOP
			}

			args := strings.Split(s.Text(), " ")
			switch args[0] {
			case "auto?":
				auto <- "get " + args[1]
				msg := <-auto
				if _, err := fmt.Fprintln(conn, msg); err != nil {
					log.Println("err:", err)
				}
			case "auto~":
				auto <- "toggle " + args[1]
				pub.pubChan <- "AUTO::" + args[1] + "::" + <-auto
			case "auto!":
				auto <- "stop " + args[1]
				pub.pubChan <- "AUTO::" + args[1] + "::false"
			case "auto":
				auto <- "start " + args[1]
				pub.pubChan <- "AUTO::" + args[1] + "::true"
			case "listen":
				pub.add(conn) // TODO: responsibility of this connection passes to publisher. Should be separate
				for k := range util.Conf.Devices {
					auto <- "get " + k
					pub.pubChan <- "AUTO::" + k + "::" + <-auto
					pubBrightness(k, pub.pubChan)
				}
				break LOOP
			case "refresh":
				pubBrightness(args[1], pub.pubChan)
			case "quit":
				// quitting implies that auto is stopped
				for k := range util.Conf.Devices {
					pub.pubChan <- "AUTO::" + k + "::false"
				}
				cancel()
				return true
			}
		}
	}

	return false
}

func daemon(ctx context.Context) error {
	// TODO: it could be beneficial to differentiate between the IPC socket and the internal one
	os.Remove(util.SOCK) // error is ignored because there is no problem if SOCK does not exist

	listener, err := net.Listen("unix", util.SOCK)
	if err != nil {
		return fmt.Errorf("daemon failed to start the listener: %w", err)
	}
	log.Println("daemon started successfully")

	var connections sync.WaitGroup
	ctxt, cancel := context.WithCancel(ctx)
	pub := newPublisher()

	var autoGroup sync.WaitGroup
	autoChan := make(chan string)
	autoGroup.Add(1)
	go func() { autoHandler(pub.pubChan, autoChan); autoGroup.Done() }()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctxt.Done():
				cancel()
				connections.Wait()
				close(autoChan)
				autoGroup.Wait()
				pub.close()
				log.Println("shutting down...")
				return nil
			default:
				log.Printf("error listening for connections: %v\n", err)
				continue
			}
		}

		connections.Add(1)
		go func() {
			if handleConn(conn, pub, autoChan, ctxt, cancel) {
				listener.Close()
			}
			connections.Done()
		}()
	}
}
